package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/join"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/end"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
)

type joinerState struct {
	games   []models.Game
	reviews []models.Review
	ends    int
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
	tx, rx := service.Run(ctx)
	joinerStateStore := store.NewStore[joinerState]()
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
				elements := msg.Elements()
				state, ok := joinerStateStore.Get(clientID)
				if !ok {
					state = joinerState{ends: 2}
				}
				if err := handleDataMessage(msg, &state, elements); err != nil {
					return err
				}
				joinerStateStore.Set(clientID, state)
			} else if msg.ExpectKind(protocol.End) {
				tx <- msg
			} else {
				return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
			}
			delivery.Ack(false)
		case msgInfo := <-rx:
			clientID := msgInfo.Options.ClientID
			state, ok := joinerStateStore.Get(clientID)
			if !ok {
				state = joinerState{ends: 2}
			}
			state.ends--
			if state.ends != 0 {
				joinerStateStore.Set(clientID, state)
				continue
			}
			slog.Debug("join and send data")
			err := joinAndSend(j, &state, msgInfo.Options)
			if err != nil {
				return err
			}
			slog.Debug("notifying coordinator")
			service.NotifyCoordinator(protocol.Reviews, msgInfo.Options)
			joinerStateStore.Delete(clientID)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func joinAndSend(j *Joiner, s *joinerState, opts protocol.MessageOptions) error {
	tuples := join.Join(s.games, s.reviews)

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

func handleDataMessage(msg protocol.Message, s *joinerState, elements *protocol.PayloadElements) error {
	if msg.HasGameData() {
		for _, element := range elements.Iter() {
			game := models.ReadGame(&element)
			s.games = append(s.games, game)
		}
	} else if msg.HasReviewData() {
		for _, element := range elements.Iter() {
			review := models.ReadReview(&element)
			s.reviews = append(s.reviews, review)
		}
	} else {
		return fmt.Errorf("unexpected data type")
	}
	return nil
}
