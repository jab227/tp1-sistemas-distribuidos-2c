package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

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
					if game.ReviewsCount < 5000 {
						continue
					}

					builder := protocol.NewPayloadBuffer(1)
					builder.BeginPayloadElement()
					data := fmt.Sprintf("%s,%s", game.AppID, game.Name)
					builder.WriteBytes([]byte(data))
					builder.EndPayloadElement()
					res := protocol.NewResultsMessage(protocol.Query4, builder.Bytes(), protocol.MessageOptions{
						MessageID: msg.GetMessageID(),
						ClientID:  msg.GetClientID(),
						RequestID: msg.GetRequestID(),
					})
					if err := r.io.Write(res.Marshal(), ""); err != nil {
						return fmt.Errorf("couldn't write query 4 output: %w", err)
					}
					slog.Debug("query 4 results", "result", data)

				}
			} else if msg.ExpectKind(protocol.End) {
				slog.Debug("received end", "node", "review_counter")
				res := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
					MessageID: msg.GetMessageID(),
					ClientID:  msg.GetClientID(),
					RequestID: msg.GetRequestID(),
				})

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
