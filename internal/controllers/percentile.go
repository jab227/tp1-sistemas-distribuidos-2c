package controllers

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"math"
	"slices"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

type innerPercentile struct {
	name    string
	counter uint
}

type percentileState map[string]innerPercentile

func (p percentileState) insertOrUpdate(game models.Game) {
	v, ok := p[game.AppID]
	if !ok {
		p[game.AppID] = innerPercentile{game.Name, 1}
		return
	}
	v.counter += 1
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

	for {
		select {
		case delivery := <-consumerCh:
			msgBytes := delivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}

			if msg.ExpectKind(protocol.Data) {
				if !msg.HasGameData() {
					return fmt.Errorf("couldn't wrong type: expected game data")
				}
				elements := msg.Elements()
				for _, element := range elements.Iter() {
					game := models.ReadGame(&element)
					r.s.insertOrUpdate(game)
				}
			} else if msg.ExpectKind(protocol.End) {
				// reset state
				slog.Debug("received end", "node", "review_counter")

				// Compute percentile
				results := make([]innerPercentile, 0, len(r.s))
				for _, v := range r.s {
					results = append(results, v)
				}
				slices.SortFunc(results, func(a, b innerPercentile) int {
					return cmp.Compare(a.counter, b.counter)
				})
				idx := percentilIndex(len(results), 90)
				// NOTE: This should be batched instead of being sent one by one
				slog.Debug("query 5 results", "result", results[idx:], "state", r.s)
				for _, result := range results[idx:] {
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

					r.s = percentileState(make(map[string]innerPercentile))
				}
				res := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
					MessageID: msg.GetMessageID(),
					ClientID:  msg.GetClientID(),
					RequestID: msg.GetRequestID(),
				})
				// Tell it ends the query 5
				res.SetQueryResult(protocol.Query5)
				if err := r.io.Write(res.Marshal(), ""); err != nil {
					return fmt.Errorf("couldn't write query 5 end: %w", err)
				}
			} else {
				return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
			}
			delivery.Ack(false)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (r *Percentile) Close() {
	r.io.Close()
}
