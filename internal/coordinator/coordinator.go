package coordinator

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

type ClientState struct {
	GamesEndCounter   int
	ReviewsEndCounter int
}

type State struct {
	state map[uint32]ClientState
}

func (s *State) AddGameEnd(clientId uint32) {
	state := s.state[clientId]
	state.GamesEndCounter++
	s.state[clientId] = state
}

func (s *State) AddReviewsEnd(clientId uint32) {
	state := s.state[clientId]
	state.ReviewsEndCounter++
	s.state[clientId] = state
}

func (s *State) GetGamesEnd(clientId uint32) int {
	state := s.state[clientId]
	return state.GamesEndCounter
}

func (s *State) GetReviewsEnd(clientId uint32) int {
	state := s.state[clientId]
	return state.ReviewsEndCounter
}

func (s *State) ResetGamesEnd(clientId uint32) {
	state := s.state[clientId]
	state.GamesEndCounter = 0
	s.state[clientId] = state
}

func (s *State) ResetReviewsEnd(clientId uint32) {
	state := s.state[clientId]
	state.ReviewsEndCounter = 0
	s.state[clientId] = state
}

type EndCoordinator struct {
	output client.OutputType
	io     client.IOManager

	expectedGamesEnd   int
	expectedReviewsEnd int

	state State

	done chan struct{}
}

func NewEndCoordinator(output client.OutputType, expectedGames int, expectedReviews int) (*EndCoordinator, error) {
	io := client.IOManager{}
	if err := io.Connect(client.InputWorker, output); err != nil {
		return nil, fmt.Errorf("error initializing IOManager %s", err)
	}

	return &EndCoordinator{
		output:             output,
		io:                 io,
		expectedGamesEnd:   expectedGames,
		expectedReviewsEnd: expectedReviews,
		state:              State{state: make(map[uint32]ClientState)},
		done:               make(chan struct{}),
	}, nil
}

func (c *EndCoordinator) Run(ctx context.Context) error {
	consumer := c.io.Input.GetConsumer()
	defer func() {
		c.done <- struct{}{}
	}()

	for {
		select {
		case delivery := <-consumer:
			bytes := delivery.Body
			msg := protocol.Message{}
			if err := msg.Unmarshal(bytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}

			if msg.ExpectKind(protocol.End) {
				if msg.HasGameData() {
					// Check if it expects games
					if c.expectedGamesEnd <= 0 {
						return fmt.Errorf("received games end but wasn't expected - Expected %d", c.expectedGamesEnd)
					}

					// Sum counter
					c.state.AddGameEnd(msg.GetClientID())
					slog.Info("Received Game END", "counter", c.state.GetGamesEnd(msg.GetClientID()))
					if c.state.GetGamesEnd(msg.GetClientID()) >= c.expectedGamesEnd {
						endMsg := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
							MessageID: msg.GetMessageID(),
							ClientID:  msg.GetClientID(),
							RequestID: msg.GetRequestID(),
						})
						<-time.After(3 * time.Second)
						slog.Info("Propagating END games", "counter", c.state.GetGamesEnd(msg.GetClientID()))
						if err := c.io.Write(endMsg.Marshal(), "game"); err != nil {
							return fmt.Errorf("couldn't write end message: %w", err)
						}
						c.state.ResetGamesEnd(msg.GetClientID())
					}

					// ACK of the MSg
					delivery.Ack(false)
				} else if msg.HasReviewData() {
					//. Check if its expects reviews
					if c.expectedReviewsEnd <= 0 {
						return fmt.Errorf("received reviews end but wasn't expected - Expected %d", c.expectedReviewsEnd)
					}

					// Sum counter
					c.state.AddReviewsEnd(msg.GetClientID())
					slog.Info("Received Review END", "counter", c.state.GetReviewsEnd(msg.GetClientID()))
					if c.state.GetReviewsEnd(msg.GetClientID()) >= c.expectedReviewsEnd {
						endMsg := protocol.NewEndMessage(protocol.Reviews, protocol.MessageOptions{
							MessageID: msg.GetMessageID(),
							ClientID:  msg.GetClientID(),
							RequestID: msg.GetRequestID(),
						})
						<-time.After(20 * time.Second)
						slog.Info("Propagating END reviews", "counter", c.state.GetReviewsEnd(msg.GetClientID()))
						if err := c.io.Write(endMsg.Marshal(), "review"); err != nil {
							return fmt.Errorf("couldn't write end message: %w", err)
						}
						c.state.ResetReviewsEnd(msg.GetClientID())
					}

					// ACK of the MSg
					delivery.Ack(false)
				} else {
					return fmt.Errorf("unexpected end type, should be game or review, got: %v", msg)
				}

			} else {
				// Should never happen
				return fmt.Errorf("unexpected message type different to END: %s", msg.GetMessageType())
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *EndCoordinator) GetDone() <-chan struct{} {
	return c.done
}

func (c *EndCoordinator) Close() {
	c.io.Close()
}
