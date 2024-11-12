package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/end"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type joinerState struct {
	games   map[string]models.Game
	reviews map[string]int
	ends    int
}

func (j *joinerState) Reset() {
	*j = joinerState{
		games:   make(map[string]models.Game),
		reviews: make(map[string]int),
		ends:    2,
	}
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
	joinerStateStore := store.NewStore[joinerState]()
	for {
		select {
		case delivery := <-ch:
			msgBytes := delivery.RecvDelivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}
			if delivery.SenderType == end.SenderPrevious {
				clientID := msg.GetClientID()
				if msg.ExpectKind(protocol.Data) {
					elements := msg.Elements()
					state, ok := joinerStateStore.Get(clientID)
					if !ok {
						state.Reset()
					}
					if err := handleDataMessage(msg, &state, elements); err != nil {
						return err
					}
					joinerStateStore.Set(clientID, state)
				} else if msg.ExpectKind(protocol.End) {
					service.NotifyNeighbours(msg)
				} else {
					return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
				}
				delivery.RecvDelivery.Ack(false)
			} else if delivery.SenderType == end.SenderNeighbour {
				clientID := msg.GetClientID()
				state, ok := joinerStateStore.Get(clientID)
				if !ok {
					state.Reset()
				}
				state.ends--
				if state.ends != 0 {
					joinerStateStore.Set(clientID, state)
					delivery.RecvDelivery.Ack(false)					
					continue
				}
				slog.Debug("join and send data")
				err := joinAndSend(j, &state, protocol.MessageOptions{
					MessageID: msg.GetMessageID(),
					ClientID:  clientID,
					RequestID: msg.GetRequestID(),
				})
				if err != nil {
					return err
				}
				slog.Debug("notifying coordinator")
				service.NotifyCoordinator(msg)
				joinerStateStore.Delete(clientID)
				delivery.RecvDelivery.Ack(false)
			} else {
				utils.Assert(false, "unknown type")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func joinAndSend(j *Joiner, s *joinerState, opts protocol.MessageOptions) error {
	for appid, count := range s.reviews {
		game := s.games[appid]
		if game.Name == "" {
			continue
		}
		builder := protocol.NewPayloadBuffer(1)
		game.BuildPayload(builder)
		for i := 0; i < count; i++ {
			res := protocol.NewDataMessage(protocol.Games, builder.Bytes(), opts)
			if err := j.io.Write(res.Marshal(), game.AppID); err != nil {
				return fmt.Errorf("couldn't write query 1 output: %w", err)
			}
		}
	}
	return nil
}

func handleDataMessage(msg protocol.Message, s *joinerState, elements *protocol.PayloadElements) error {
	if msg.HasGameData() {
		for _, element := range elements.Iter() {
			game := models.ReadGame(&element)
			s.games[game.AppID] = game
		}
	} else if msg.HasReviewData() {
		for _, element := range elements.Iter() {
			review := models.ReadReview(&element)
			s.reviews[review.AppID] += 1
		}
	} else {
		return fmt.Errorf("unexpected data type")
	}
	return nil
}
