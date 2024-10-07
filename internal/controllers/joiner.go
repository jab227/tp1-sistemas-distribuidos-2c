package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/join"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

type joinerState struct {
	games   []models.Game
	reviews []models.Review
	ends    int
}

type Joiner struct {
	io   client.IOManager
	done chan struct{}
	s    joinerState
}

func NewJoiner() (*Joiner, error) {
	var io client.IOManager
	// TODO: Change to the correct configs
	// TODO: Check ouput type by env
	if err := io.Connect(client.InputWorker, client.OutputWorker); err != nil {
		return nil, fmt.Errorf("couldn't create os counter: %w", err)
	}
	return &Joiner{
		io:   io,
		done: make(chan struct{}),
		s:    joinerState{ends: 2},
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

	for {
		select {
		case delivery := <-consumerCh:
			msgBytes := delivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}
			if msg.ExpectKind(protocol.Data) {
				elements := msg.Elements()
				if err := j.handleDataMessage(msg, elements); err != nil {
					return err
				}
			} else if msg.ExpectKind(protocol.End) {
				// reset state
				slog.Debug("received end", "node", "joiner")
				j.s.ends--
				if j.s.ends != 0 {
					continue
				}

				tuples := join.Join(j.s.games, j.s.reviews)
				// NOTE(juan): This would be more
				// efficient with batching but for now
				// it's okay
				for _, tuple := range tuples {
					game := tuple.X
					builder := protocol.NewPayloadBuffer(1)
					game.BuildPayload(builder)
					res := protocol.NewDataMessage(protocol.Games, builder.Bytes(), protocol.MessageOptions{
						MessageID: msg.GetMessageID(),
						ClientID:  msg.GetClientID(),
						RequestID: msg.GetRequestID(),
					})
					if err := j.io.Write(res.Marshal(), ""); err != nil {
						return fmt.Errorf("couldn't write query 1 output: %w", err)
					}
				}
				// TODO: Send end and handle sync
				j.s = joinerState{}
			} else {
				return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
			}
			delivery.Ack(false)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (j *Joiner) handleDataMessage(msg protocol.Message, elements *protocol.PayloadElements) error {
	if msg.HasGameData() {
		for _, element := range elements.Iter() {
			game := models.ReadGame(&element)
			j.s.games = append(j.s.games, game)
		}
	} else if msg.HasReviewData() {
		for _, element := range elements.Iter() {
			review := models.ReadReview(&element)
			j.s.reviews = append(j.s.reviews, review)
		}
	} else {
		return fmt.Errorf("unexpected data type")
	}
	return nil
}
