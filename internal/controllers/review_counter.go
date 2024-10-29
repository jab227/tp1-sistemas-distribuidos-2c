package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

type inner struct {
	name    string
	counter uint
}
type reviewCounterState map[string]inner

func (r reviewCounterState) insertOrUpdate(game models.Game) {
	v, ok := r[game.AppID]
	if !ok {
		r[game.AppID] = inner{game.Name, 1}
		return
	}
	v.counter += 1
	r[game.AppID] = v
}

type ReviewCounter struct {
	io   client.IOManager
	done chan struct{}
	s    reviewCounterState
}

func NewReviewCounter() (*ReviewCounter, error) {
	var io client.IOManager
	if err := io.Connect(client.DirectSubscriber, client.OutputWorker); err != nil {
		return nil, fmt.Errorf("couldn't create os counter: %w", err)
	}
	return &ReviewCounter{
		io:   io,
		done: make(chan struct{}),
		s:    reviewCounterState(make(map[string]inner)),
	}, nil
}

func (r *ReviewCounter) Destroy() {
	r.io.Close()
}

func (r *ReviewCounter) Done() <-chan struct{} {
	return r.done
}

func (r *ReviewCounter) Run(ctx context.Context) error {
	consumerCh := r.io.Input.GetConsumer()
	defer func() {
		r.done <- struct{}{}
	}()
	end := false
	for {
		select {
		case delivery := <-consumerCh:
			if end {
				slog.Debug("received message after end")
			}
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
				end = true
				slog.Debug("received end", "node", "review_counter")
				// NOTE: This should be batched instead of being sent one by one
				for _, result := range r.s {
					if result.counter < 5000 {
						continue
					}
					builder := protocol.NewPayloadBuffer(1)
					builder.BeginPayloadElement()
					builder.WriteBytes([]byte(result.name))
					builder.EndPayloadElement()
					res := protocol.NewResultsMessage(protocol.Query4, builder.Bytes(), protocol.MessageOptions{
						MessageID: msg.GetMessageID(),
						ClientID:  msg.GetClientID(),
						RequestID: msg.GetRequestID(),
					})
					if err := r.io.Write(res.Marshal(), ""); err != nil {
						return fmt.Errorf("couldn't write query 4 output: %w", err)
					}
					slog.Debug("query 4 results", "result", result.name)
				}
				r.s = reviewCounterState(make(map[string]inner))				
				res := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
					MessageID: msg.GetMessageID(),
					ClientID:  msg.GetClientID(),
					RequestID: msg.GetRequestID(),
				})
				// Tell it ends the query 4
				res.SetQueryResult(protocol.Query4)
				if err := r.io.Write(res.Marshal(), ""); err != nil {
					return fmt.Errorf("couldn't write query 4 end: %w", err)
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

func (r *ReviewCounter) Close() {
	r.io.Close()
}
