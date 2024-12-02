package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/heap"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/batch"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/persistence"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type topReviewsState struct {
	appByReviewScore map[string]int
}

type TopReviews struct {
	iomanager client.IOManager
	done      chan struct{}
	n         int
}

func NewTopReviews(n int) (*TopReviews, error) {
	ioManager := client.IOManager{}
	err := ioManager.Connect(client.DirectSubscriber, client.OutputWorker)
	if err != nil {
		return nil, err
	}

	return &TopReviews{
		iomanager: ioManager,
		done:      make(chan struct{}, 1),
		n:         n,
	}, nil
}

func (tr *TopReviews) Done() <-chan struct{} {
	return tr.done
}

func (tr *TopReviews) Run(ctx context.Context) error {
	consumerCh := tr.iomanager.Input.GetConsumer()
	defer func() {
		tr.done <- struct{}{}
	}()

	topReviewsStateStore, log, err := reloadTopReviews(tr)
	if err != nil {
		return err
	}
	batcher := batch.NewBatcher(3000)
	for {
		select {
		case delivery := <-consumerCh:
			bytes := delivery.Body
			msg := protocol.Message{}
			if err := msg.Unmarshal(bytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}

			batcher.Push(msg, delivery)
			if !batcher.IsFull() {
				continue
			}
			currentBatch := batcher.Batch()
			log.Append(batch.MarshalBatch(currentBatch), uint32(TXNBatch))

			if err := processTopReviewsBatch(currentBatch, topReviewsStateStore, tr); err != nil {
				return err
			}

			if err := log.Commit(); err != nil {
				return fmt.Errorf("couldn't commit to disk: %w", err)
			}
			time.Sleep(5 * time.Second)

			batcher.Acknowledge()
		case <-time.After(10 * time.Second):
			if batcher.IsEmpty() {
				continue
			}
			currentBatch := batcher.Batch()
			log.Append(batch.MarshalBatch(currentBatch), uint32(TXNBatch))

			if err := processTopReviewsBatch(currentBatch, topReviewsStateStore, tr); err != nil {
				return err
			}

			if err := log.Commit(); err != nil {
				return fmt.Errorf("couldn't commit to disk: %w", err)
			}
			time.Sleep(5 * time.Second)

			batcher.Acknowledge()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func reloadTopReviews(tr *TopReviews) (store.Store[*topReviewsState], *persistence.TransactionLog, error) {
	stateStore := store.NewStore[*topReviewsState]()
	log := persistence.NewTransactionLog("../logs/top_reviews.log")
	logBytes, err := os.ReadFile("../logs/top_reviews.log")
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
		case TXNBatch:
			currentBatch, err := batch.UnmarshalBatch(entry.Data)
			if err != nil {
				return stateStore, log, err
			}
			applyTopReviewsBatch(currentBatch, stateStore, tr)
		}
	}
	return stateStore, log, nil
}

func applyTopReviewsBatch(
	batch []protocol.Message,
	topReviewsStateStore store.Store[*topReviewsState],
	tr *TopReviews,
) {
	for _, msg := range batch {
		if msg.ExpectKind(protocol.Data) {
			utils.Assert(msg.HasGameData(), "wrong type: expected game data")
			clientID := msg.GetClientID()
			state, ok := topReviewsStateStore.Get(clientID)
			if !ok {
				state = &topReviewsState{make(map[string]int)}
				topReviewsStateStore.Set(clientID, state)
			}
			tr.processReviewsData(state, msg)
		} else if msg.ExpectKind(protocol.End) {
			continue
		} else {
			utils.Assertf(false, "unexpected message type: %s", msg.GetMessageType())
		}
	}
}

func processTopReviewsBatch(batch []protocol.Message, topReviewsStateStore store.Store[*topReviewsState], tr *TopReviews) error {
	for _, msg := range batch {
		clientID := msg.GetClientID()
		state, ok := topReviewsStateStore.Get(clientID)
		if !ok {
			state = &topReviewsState{make(map[string]int)}
			topReviewsStateStore.Set(clientID, state)
		}

		if msg.ExpectKind(protocol.Data) {
			if !msg.HasGameData() {
				return fmt.Errorf("wrong type: expected game data")
			}
			tr.processReviewsData(state, msg)
		} else if msg.ExpectKind(protocol.End) {
			slog.Debug("received end", "game", msg.HasGameData())
			if err := tr.writeResult(state, msg); err != nil {
				return err
			}
			topReviewsStateStore.Delete(clientID)
		} else {
			return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
		}
	}
	return nil
}

func (tr *TopReviews) processReviewsData(state *topReviewsState, internalMsg protocol.Message) {
	elements := internalMsg.Elements()
	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		key := fmt.Sprintf("%s||%s", game.AppID, game.Name)
		state.appByReviewScore[key] = int(game.ReviewsCount)
	}
}

func (tr *TopReviews) writeResult(state *topReviewsState, internalMsg protocol.Message) error {
	values := make([]heap.Value, 0, len(state.appByReviewScore))
	for k, v := range state.appByReviewScore {
		name := strings.Split(k, "||")[1]
		count := v
		values = append(values, heap.Value{Name: name, Count: count})
	}
	h := heap.NewHeap()
	for _, value := range values {
		h.PushValue(value)
	}

	results := h.TopN(int(tr.n))
	slog.Debug("topn results", "results", results)
	for _, value := range results {
		buffer := protocol.NewPayloadBuffer(1)
		buffer.BeginPayloadElement()
		buffer.WriteBytes([]byte(value.Name))
		buffer.EndPayloadElement()
		response := protocol.NewResultsMessage(protocol.Query3, buffer.Bytes(), protocol.MessageOptions{
			MessageID: internalMsg.GetMessageID(),
			ClientID:  internalMsg.GetClientID(),
			RequestID: internalMsg.GetRequestID(),
		})
		if err := tr.iomanager.Write(response.Marshal(), ""); err != nil {
			return fmt.Errorf("couldn't write query 3 output: %w", err)
		}
	}
	res := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
		MessageID: internalMsg.GetMessageID(),
		ClientID:  internalMsg.GetClientID(),
		RequestID: internalMsg.GetRequestID(),
	})
	// Tell it ends the query 3
	res.SetQueryResult(protocol.Query3)
	if err := tr.iomanager.Write(res.Marshal(), ""); err != nil {
		return fmt.Errorf("couldn't write query 5 end: %w", err)
	}
	slog.Debug("query 3 results", "result", res, "state", state)

	return nil
}

func (t *TopReviews) Close() {
	t.iomanager.Close()
}
