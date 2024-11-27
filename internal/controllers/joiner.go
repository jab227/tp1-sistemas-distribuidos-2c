package controllers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/batch"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/end"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/persistence"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type EndSet byte
type EndType byte

const (
	EndTypeGames   EndType = 0b01
	EndTypeReviews EndType = 0b10
	EndTypeAll     EndType = 0b11
)

func (e *EndSet) Set(t EndType) {
	*e |= EndSet(t)
}

func (e EndSet) HasBoth() bool {
	return e == EndSet(EndTypeAll)
}

type joinerState struct {
	Games   map[string]models.Game
	Reviews map[string]map[uint32]struct{}
	Ends    EndSet
}

func (j *joinerState) Reset() {
	*j = joinerState{
		Games:   make(map[string]models.Game),
		Reviews: make(map[string]map[uint32]struct{}),
	}
}

func MarshalStore(s store.Store[joinerState]) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(s); err != nil {
		return nil, err
	}

	var compressedBuffer bytes.Buffer
	w, err := gzip.NewWriterLevel(&compressedBuffer, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	io.Copy(w, &buf)
	w.Close()
	return compressedBuffer.Bytes(), nil
}

func UnmarshalStore(p []byte) (store.Store[joinerState], error) {
	buf := bytes.NewBuffer(p)
	r, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	decoder := gob.NewDecoder(r)
	s := store.NewStore[joinerState]()
	if err := decoder.Decode(&s); err != nil {
		return nil, err
	}
	return s, nil
}

type Joiner struct {
	io   client.IOManager
	done chan struct{}
}

func NewJoiner() (*Joiner, error) {
	var io client.IOManager
	if err := io.Connect(client.DirectSubscriber, client.Router); err != nil {
		return nil, fmt.Errorf("couldn't create os counter: %w", err)
	}
	return &Joiner{
		io:   io,
		done: make(chan struct{}),
	}, nil
}

func (j *Joiner) Destroy() {
	j.io.Close()
}

func (j *Joiner) Done() <-chan struct{} {
	return j.done
}

func applyJoinerBatch(
	batch []protocol.Message,
	joinerStateStore store.Store[joinerState],
) error {

	for _, m := range batch {
		clientID := m.GetClientID()
		if m.ExpectKind(protocol.Data) {
			elements := m.Elements()
			state, ok := joinerStateStore.Get(clientID)
			if !ok {
				state.Reset()
			}
			if err := handleDataMessage(m, &state, elements); err != nil {
				return err
			}
			joinerStateStore.Set(clientID, state)
		} else if m.ExpectKind(protocol.End) {
			continue
		} else {
			utils.Assertf(false, "unexpected message type: %s", m.GetMessageType())
		}
	}
	return nil
}

func reloadJoiner() (store.Store[joinerState], *persistence.TransactionLog, error) {
	stateStore := store.NewStore[joinerState]()
	idptr, err := utils.GetFromEnv("NODE_ID")
	if err != nil {
		panic("NODE_ID should be set")
	}
	id := *idptr
	logFile := fmt.Sprintf("../logs/joiner-%s.log", id)
	log := persistence.NewTransactionLog(logFile)
	logBytes, err := os.ReadFile(logFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return stateStore, log, nil
		}
		return stateStore, log, err
	}
	if err := log.Unmarshal(logBytes); err != nil {
		err = fmt.Errorf("couldn't unmarshal log: %w", err)
		return stateStore, log, err
	}
	for _, entry := range log.GetLog() {
		switch TXN(entry.Kind) {
		case TXNEnd:
			var endMessage protocol.Message
			if err := endMessage.Unmarshal(entry.Data); err != nil {
				return stateStore, log, err
			}
			utils.Assert(endMessage.ExpectKind(protocol.End), "expected end message")
			clientID := endMessage.GetClientID()
			state, ok := stateStore.Get(clientID)
			if !ok {
				state.Reset()
			}
			state.Ends.Set(chooseEndType(endMessage))
			// if the second end is already in disk it
			// means that the data was already sent and we
			// can delete the client from the store
			if state.Ends.HasBoth() {
				stateStore.Delete(endMessage.GetClientID())
			}
		case TXNBatch:
			stateStore, err = UnmarshalStore(entry.Data)
			if err != nil {
				return stateStore, log, err
			}
		}
	}
	return stateStore, log, nil
}

func (j *Joiner) Run(ctx context.Context) error {
	consumerCh := j.io.Input.GetConsumer()
	defer func() {
		j.done <- struct{}{}
	}()
	options, err := end.GetServiceOptionsFromEnv()
	if err != nil {
		return err
	}
	service, err := end.NewService(options)
	ch := service.MergeConsumers(consumerCh)
	joinerStateStore, log, err := reloadJoiner()
	if err != nil {
		return err
	}
	batcher := batch.NewBatcher(5000)
	for {
		select {
		case delivery := <-ch:
			body := delivery.RecvDelivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(body); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}
			if delivery.SenderType == end.SenderPrevious {
				batcher.Push(msg, delivery.RecvDelivery)
				if !batcher.IsFull() {
					continue
				}
				currentBatch := batcher.Batch()
				storeData, err := MarshalStore(joinerStateStore)
				if err != nil {
					return err
				}
				log.Append(storeData, uint32(TXNBatch))

				if err := processBatch(currentBatch, joinerStateStore, service); err != nil {
					return err
				}
				if err := log.Commit(); err != nil {
					return fmt.Errorf("couldn't commit to disk: %w", err)
				}
				batcher.Acknowledge()
			} else if delivery.SenderType == end.SenderNeighbour {
				clientID := msg.GetClientID()
				log.Append(msg.Marshal(), uint32(TXNEnd))
				state, ok := joinerStateStore.Get(clientID)
				if !ok {
					state.Reset()
				}
				state.Ends.Set(chooseEndType(msg))
				if !state.Ends.HasBoth() {
					if err := log.Commit(); err != nil {
						return fmt.Errorf("couldn't commit to disk: %w", err)
					}
					joinerStateStore.Set(clientID, state)
					delivery.RecvDelivery.Ack(false)
					continue
				}
				slog.Debug("join and send data")
				err := joinAndSend(j, &state, protocol.MessageOptions{
					ClientID:  clientID,
					RequestID: msg.GetRequestID(),
				})
				if err != nil {
					return err
				}
				slog.Debug("notifying coordinator")
				msg = protocol.NewEndMessage(protocol.Reviews, protocol.MessageOptions{
					MessageID: msg.GetMessageID(),
					ClientID:  clientID,
					RequestID: msg.GetRequestID(),
				})
				service.NotifyCoordinator(msg)
				joinerStateStore.Delete(clientID)
				if err := log.Commit(); err != nil {
					return fmt.Errorf("couldn't commit to disk: %w", err)
				}
				delivery.RecvDelivery.Ack(false)
			} else {
				utils.Assert(false, "unknown type")
			}
		case <-time.After(10 * time.Second):
			if batcher.IsEmpty() {
				continue
			}
			currentBatch := batcher.Batch()
			storeData, err := MarshalStore(joinerStateStore)
			if err != nil {
				return err
			}
			log.Append(storeData, uint32(TXNBatch))
			if err := processBatch(currentBatch, joinerStateStore, service); err != nil {
				return err
			}
			if err := log.Commit(); err != nil {
				return fmt.Errorf("couldn't commit to disk: %w", err)
			}
			batcher.Acknowledge()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func processBatch(batch []protocol.Message, joinerStateStore store.Store[joinerState], service *end.Service) error {
	for _, m := range batch {
		clientID := m.GetClientID()
		if m.ExpectKind(protocol.Data) {
			elements := m.Elements()
			state, ok := joinerStateStore.Get(clientID)
			if !ok {
				state.Reset()
			}
			if err := handleDataMessage(m, &state, elements); err != nil {
				return err
			}
			joinerStateStore.Set(clientID, state)
		} else if m.ExpectKind(protocol.End) {
			service.NotifyNeighbours(m)
		} else {
			return fmt.Errorf("unexpected message type: %s", m.GetMessageType())
		}
	}
	return nil
}

func joinAndSend(j *Joiner, s *joinerState, opts protocol.MessageOptions) error {
	for appid, set := range s.Reviews {
		game := s.Games[appid]
		if game.Name == "" {
			continue
		}
		game.ReviewsCount = uint32(len(set))
		builder := protocol.NewPayloadBuffer(1)
		game.BuildPayload(builder)

		slog.Info("Message to send", "msg", fmt.Sprintf("%v", game))
		res := protocol.NewDataMessage(protocol.Games, builder.Bytes(), opts)
		if err := j.io.Write(res.Marshal(), game.AppID); err != nil {
			return fmt.Errorf("couldn't write query 1 output: %w", err)
		}
	}
	return nil
}

func handleDataMessage(msg protocol.Message, s *joinerState, elements *protocol.PayloadElements) error {
	if msg.HasGameData() {
		for _, element := range elements.Iter() {
			game := models.ReadGame(&element)
			s.Games[game.AppID] = game
		}
	} else if msg.HasReviewData() {
		for _, element := range elements.Iter() {
			review := models.ReadReview(&element)
			v, ok := s.Reviews[review.AppID]
			if !ok {
				v = make(map[uint32]struct{})
				s.Reviews[review.AppID] = v
			}
			v[msg.GetMessageID()] = struct{}{}
		}
	} else {
		return fmt.Errorf("unexpected data type")
	}
	return nil
}

func chooseEndType(msg protocol.Message) EndType {
	if msg.HasGameData() {
		return EndTypeGames
	} else if msg.HasReviewData() {
		return EndTypeReviews
	} else {
		panic("shouldn't happen because there isn't another data type")
	}
}
