package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
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
}

func NewReviewCounter() (*ReviewCounter, error) {
	var io client.IOManager
	if err := io.Connect(client.DirectSubscriber, client.OutputWorker); err != nil {
		return nil, fmt.Errorf("couldn't create os counter: %w", err)
	}
	return &ReviewCounter{
		io:   io,
		done: make(chan struct{}),
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
	reviewCounterStateStore := store.NewStore[reviewCounterState]()
	for {
		select {
		case delivery := <-consumerCh:
			msgBytes := delivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}
			clientID := msg.GetClientID()
			if msg.ExpectKind(protocol.Data) {
				if !msg.HasGameData() {
					return fmt.Errorf("couldn't wrong type: expected game data")
				}
				elements := msg.Elements()
				state, ok := reviewCounterStateStore.Get(clientID)
				if !ok {
					// the map behaves like a ptr
					state = make(map[string]inner)
					reviewCounterStateStore.Set(clientID, state)
				}
				for _, element := range elements.Iter() {
					game := models.ReadGame(&element)
					state.insertOrUpdate(game)
				}
			} else if msg.ExpectKind(protocol.End) {
				slog.Debug("received end", "node", "review_counter")
				// It should exist or if doesn't exist we send empty results
				state, _ := reviewCounterStateStore.Get(clientID)
				if err := marshalAndSendResults(r, state, msg); err != nil {
					return err
				}
				reviewCounterStateStore.Delete(clientID)
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

func marshalAndSendResults(r *ReviewCounter, s reviewCounterState, msg protocol.Message) error {
	for _, result := range s {
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
	res := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
		MessageID: msg.GetMessageID(),
		ClientID:  msg.GetClientID(),
		RequestID: msg.GetRequestID(),
	})

	res.SetQueryResult(protocol.Query4)
	if err := r.io.Write(res.Marshal(), ""); err != nil {
		return fmt.Errorf("couldn't write query 4 end: %w", err)
	}
	return nil
}

func (r *ReviewCounter) Close() {
	r.io.Close()
}
