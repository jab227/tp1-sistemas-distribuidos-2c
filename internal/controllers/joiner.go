package controllers

import (
	"context"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/join"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/end"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"log/slog"
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
	if err := io.Connect(client.DirectSubscriber, client.Router); err != nil {
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
	options, err := end.GetServiceOptionsFromEnv()
	if err != nil {
		return err
	}
	service, err := end.NewRouterService(options)
	if err != nil {
		return err
	}
	service.Run(ctx)
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
				elements := msg.Elements()
				if err := j.handleDataMessage(msg, elements); err != nil {
					return err
				}
			} else if msg.ExpectKind(protocol.End) {
				j.s.ends--
				if j.s.ends != 0 {
					continue
				}
				end = true
				slog.Debug("join and send data")
				options := protocol.MessageOptions{
					MessageID: msg.GetMessageID(),
					ClientID:  msg.GetClientID(),
					RequestID: msg.GetRequestID(),
				}
				err := joinAndSend(j, options)
				if err != nil {
					return err
				}
				slog.Debug("notifying coordinator")
				service.NotifyCoordinator(protocol.Reviews, options)
			} else {
				return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
			}
			delivery.Ack(false)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func joinAndSend(j *Joiner, opts protocol.MessageOptions) error {
	tuples := join.Join(j.s.games, j.s.reviews)

	for _, tuple := range tuples {
		game := tuple.X
		builder := protocol.NewPayloadBuffer(1)
		game.BuildPayload(builder)
		res := protocol.NewDataMessage(protocol.Games, builder.Bytes(), opts)
		if err := j.io.Write(res.Marshal(), game.AppID); err != nil {
			return fmt.Errorf("couldn't write query 1 output: %w", err)
		}
	}
	return nil
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
