package controllers

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/batch"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
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
	v.counter += uint(game.ReviewsCount)
	p[game.AppID] = v
}

func percentilIndex(length, percentil int) int {
	return int(math.Round(float64(length) * float64(percentil) / 100.0))
}

type Percentile struct {
	io   client.IOManager
	done chan struct{}
	s    percentileState
}

func NewPercentile() (*Percentile, error) {
	var io client.IOManager
	if err := io.Connect(client.DirectSubscriber, client.OutputWorker); err != nil {
		return nil, fmt.Errorf("couldn't create os counter: %w", err)
	}
	return &Percentile{
		io:   io,
		done: make(chan struct{}),
		s:    percentileState(make(map[string]innerPercentile)),
	}, nil
}

func (r *Percentile) Destroy() {
	r.io.Close()
}

func (r *Percentile) Done() <-chan struct{} {
	return r.done
}

func (r *Percentile) Run(ctx context.Context) error {
	consumerCh := r.io.Input.GetConsumer()
	defer func() {
		r.done <- struct{}{}
	}()

	percentileStateStore := store.NewStore[percentileState]()
	batcher := batch.NewBatcher(2000)
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
			batch := batcher.Batch()
			if err := processPercentileBatch(batch, percentileStateStore, r); err != nil {
				return err
			}
			batcher.Acknowledge()
		case <-time.After(10 * time.Second):
			if batcher.IsEmpty() {
				continue
			}
			batch := batcher.Batch()
			if err := processPercentileBatch(batch, percentileStateStore, r); err != nil {
				return err
			}
			batcher.Acknowledge()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func processPercentileBatch(batch []protocol.Message, percentileStateStore store.Store[percentileState], r *Percentile) error {
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

			slog.Debug("query 5 results", "result", results, "state", s)

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

func marshalAndSendPercentileResults(results []innerPercentile, msg protocol.Message, r *Percentile) error {
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
