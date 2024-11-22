package controllers

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"slices"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/batch"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/persistence"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type innerPercentile struct {
	name    string
	counter uint
}

type percentileState map[string]innerPercentile

func (p percentileState) insertOrUpdate(game models.Game) {
	v, ok := p[game.AppID]
	if !ok {
		p[game.AppID] = innerPercentile{game.Name, uint(game.ReviewsCount)}
		return
	}
	v.counter = uint(game.ReviewsCount)
	p[game.AppID] = v
}

func percentilIndex(length, percentil int) int {
	return int(math.Round(float64(length) * float64(percentil) / 100.0))
}

type Percentile struct {
	io   client.IOManager
	done chan struct{}
}

func NewPercentile() (*Percentile, error) {
	var io client.IOManager
	if err := io.Connect(client.DirectSubscriber, client.OutputWorker); err != nil {
		return nil, fmt.Errorf("couldn't create os counter: %w", err)
	}
	return &Percentile{
		io:   io,
		done: make(chan struct{}),
	}, nil
}

func (r *Percentile) Destroy() {
	r.io.Close()
}

func (r *Percentile) Done() <-chan struct{} {
	return r.done
}

func reloadPercentile() (store.Store[percentileState], *persistence.TransactionLog,  error) {
	stateStore := store.NewStore[percentileState]()
	log := persistence.NewTransactionLog("../logs/percentile.log")
	logBytes, err := os.ReadFile("../logs/percentile.log")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return stateStore, log, nil
		}
		return stateStore, log, err
	}
	if err := log.Unmarshal(logBytes); err != nil {
		err = fmt.Errorf("couldn't unmarshal log: %w", err)
		return stateStore, log,  err
	}
	for _, entry := range log.GetLog() {
		switch TXN(entry.Kind) {
		case TXNBatch:
			currentBatch, err := batch.UnmarshalBatch(entry.Data)
			if err != nil {
				return stateStore, log, err
			}
			applyPercentileBatch(currentBatch, stateStore)
		}
	}
	return stateStore, log,  nil
}

func (r *Percentile) Run(ctx context.Context) error {
	consumerCh := r.io.Input.GetConsumer()
	defer func() {
		r.done <- struct{}{}
	}()

	percentileStateStore, log, err := reloadPercentile()
	if err != nil {
		return err
	}
	batcher := batch.NewBatcher(50)
	for {
		select {
		case delivery := <-consumerCh:
			msgBytes := delivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}
			batcher.Push(msg, delivery)
			if !batcher.IsFull() {
				continue
			}

			currentBatch := batcher.Batch()
			log.Append(batch.MarshalBatch(currentBatch), uint32(TXNBatch))
			slog.Debug("processing batch")
			if err := processPercentileBatch(
				currentBatch,
				percentileStateStore,
				r,
			); err != nil {
				return err
			}
			slog.Debug("commit")
			if err := log.Commit(); err != nil {
				return fmt.Errorf("couldn't commit to disk: %w", err)
			}
			time.Sleep(5 * time.Second)
			slog.Debug("acknowledge")
			batcher.Acknowledge()
		case <-time.After(10 * time.Second):
			if batcher.IsEmpty() {
				continue
			}
			currentBatch := batcher.Batch()
			log.Append(batch.MarshalBatch(currentBatch), uint32(TXNBatch))
			slog.Debug("processing batch timeout")
			if err := processPercentileBatch(
				currentBatch,
				percentileStateStore,
				r,
			); err != nil {
				return err
			}
			slog.Debug("storing message id batch timeout")
			slog.Debug("commit batch timeout")
			if err := log.Commit(); err != nil {
				return fmt.Errorf("couldn't commit to disk: %w", err)
			}
			slog.Debug("acknowledge batch timeout")
			time.Sleep(5 * time.Second)
			batcher.Acknowledge()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func processPercentileBatch(
	batch []protocol.Message,
	percentileStateStore store.Store[percentileState],
	r *Percentile,
) error {
	for _, msg := range batch {
		if msg.ExpectKind(protocol.Data) {
			if !msg.HasGameData() {
				return fmt.Errorf("couldn't wrong type: expected game data")
			}
			clientID := msg.GetClientID()
			elements := msg.Elements()
			state, ok := percentileStateStore.Get(clientID)
			if !ok {
				state = make(map[string]innerPercentile)
				percentileStateStore.Set(clientID, state)
			}
			for _, element := range elements.Iter() {
				game := models.ReadGame(&element)
				state.insertOrUpdate(game)
			}
		} else if msg.ExpectKind(protocol.End) {
			clientID := msg.GetClientID()

			slog.Debug("received end", "node", "percentile", "game end", msg.HasGameData(), "review end", msg.HasReviewData())

			s, _ := percentileStateStore.Get(clientID)
			results := computePercentile(s)

			slog.Debug("query 5 results", "result", results)

			if err := marshalAndSendPercentileResults(results, msg, r); err != nil {
				return err
			}
			percentileStateStore.Delete(clientID)
		} else {
			return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
		}
	}
	return nil
}

func applyPercentileBatch(
	batch []protocol.Message,
	percentileStateStore store.Store[percentileState],
) {
	for _, msg := range batch {
		if msg.ExpectKind(protocol.Data) {
			utils.Assert(msg.HasGameData(), "wrong type: expected game data")
			clientID := msg.GetClientID()
			elements := msg.Elements()
			state, ok := percentileStateStore.Get(clientID)
			if !ok {
				state = make(map[string]innerPercentile)
				percentileStateStore.Set(clientID, state)
			}
			for _, element := range elements.Iter() {
				game := models.ReadGame(&element)
				state.insertOrUpdate(game)
			}
		} else if msg.ExpectKind(protocol.End) {
			continue
		} else {
			utils.Assertf(false, "unexpected message type: %s", msg.GetMessageType())
		}
	}
}

func marshalAndSendPercentileResults(results []innerPercentile, msg protocol.Message, r *Percentile) error {
	slog.Debug("message result", "clientID", msg.GetClientID())
	for _, result := range results {
		builder := protocol.NewPayloadBuffer(1)
		builder.BeginPayloadElement()
		builder.WriteBytes([]byte(result.name))
		builder.EndPayloadElement()
		res := protocol.NewResultsMessage(protocol.Query5, builder.Bytes(), protocol.MessageOptions{
			MessageID: msg.GetMessageID(),
			ClientID:  msg.GetClientID(),
			RequestID: msg.GetRequestID(),
		})
		if err := r.io.Write(res.Marshal(), ""); err != nil {
			return fmt.Errorf("couldn't write query 5 output: %w", err)
		}

	}
	// Generate own messageID
	res := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
		MessageID: msg.GetMessageID(),
		ClientID:  msg.GetClientID(),
		RequestID: msg.GetRequestID(),
	})

	res.SetQueryResult(protocol.Query5)
	if err := r.io.Write(res.Marshal(), ""); err != nil {
		return fmt.Errorf("couldn't write query 5 end: %w", err)
	}
	return nil
}

func computePercentile(s percentileState) []innerPercentile {
	results := make([]innerPercentile, 0, len(s))
	for _, v := range s {
		results = append(results, v)
	}
	slices.SortFunc(results, func(a, b innerPercentile) int {
		return cmp.Compare(a.counter, b.counter)
	})
	idx := percentilIndex(len(results), 90)
	return results[idx:]
}

func (r *Percentile) Close() {
	r.io.Close()
}
