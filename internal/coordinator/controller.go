package coordinator

import (
	"context"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"log/slog"
	"time"
)

type EndState struct {
	Games   map[uint32]map[uint32]struct{} // Map with nodes ID, each with a set of clientIds
	Reviews map[uint32]map[uint32]struct{} // Map with nodes ID, each with a set of clientsId

	SentGames   map[uint32]struct{}
	SentReviews map[uint32]struct{}
}

func NewEndState(numberOfNodes uint32) *EndState {
	games := make(map[uint32]map[uint32]struct{}, numberOfNodes)
	reviews := make(map[uint32]map[uint32]struct{}, numberOfNodes)

	for i := uint32(1); i <= numberOfNodes; i++ {
		games[i] = make(map[uint32]struct{})
		reviews[i] = make(map[uint32]struct{})
	}

	return &EndState{
		Games:       games,
		Reviews:     reviews,
		SentGames:   make(map[uint32]struct{}),
		SentReviews: make(map[uint32]struct{}),
	}
}

// TODO - Delete
func (s *EndState) PrintState() {
	slog.Info("State", "games", s.Games, "reviews", s.Reviews, "sentGames", s.SentGames, "sentReviews", s.SentReviews)
}

func (s *EndState) WasSentGame(clientId uint32) bool {
	_, ok := s.SentGames[clientId]
	return ok
}

func (s *EndState) WasSentReview(clientId uint32) bool {
	_, ok := s.SentReviews[clientId]
	return ok
}

func (s *EndState) SetGameSent(clientId uint32) {
	if _, ok := s.SentGames[clientId]; !ok {
		s.SentGames[clientId] = struct{}{}
	}
	s.SentGames[clientId] = struct{}{}
}

func (s *EndState) SetReviewSent(clientId uint32) {
	if _, ok := s.SentReviews[clientId]; !ok {
		s.SentReviews[clientId] = struct{}{}
	}
	s.SentReviews[clientId] = struct{}{}
}

func (s *EndState) AddGame(nodeId uint32, clientId uint32) {
	nodeState, ok := s.Games[nodeId]
	if !ok {
		nodeState = make(map[uint32]struct{})
	}

	v, ok := nodeState[clientId]
	if !ok {
		v = struct{}{}
		nodeState[clientId] = v
	}

	s.Games[nodeId] = nodeState
}

func (s *EndState) AddReview(nodeId uint32, clientId uint32) {
	nodeState, ok := s.Reviews[nodeId]
	if !ok {
		nodeState = make(map[uint32]struct{})
	}

	v, ok := nodeState[clientId]
	if !ok {
		v = struct{}{}
		nodeState[clientId] = v
	}

	s.Reviews[nodeId] = nodeState
}

func (s *EndState) AllGamesForId(clientId uint32) bool {
	for _, value := range s.Games {
		if _, ok := value[clientId]; !ok {
			return false
		}
	}

	return true
}

func (s *EndState) AllReviewsForId(clientId uint32) bool {
	for _, value := range s.Reviews {
		if _, ok := value[clientId]; !ok {
			return false
		}
	}

	return true
}

func (s *EndState) ResetGame(clientId uint32) {
	for _, value := range s.Games {
		delete(value, clientId)
	}
}

func (s *EndState) ResetReview(clientId uint32) {
	for _, value := range s.Reviews {
		delete(value, clientId)
	}
}

type EndCoordinatorController struct {
	output client.OutputType
	io     client.IOManager

	state EndState

	done chan struct{}
}

func NewEndCoordinatorController(output client.OutputType, numberOfNodes uint32) (*EndCoordinatorController, error) {
	io := client.IOManager{}
	if err := io.Connect(client.InputWorker, output); err != nil {
		return nil, fmt.Errorf("error initializing IOManager %s", err)
	}

	return &EndCoordinatorController{
		output: output,
		io:     io,
		state:  *NewEndState(numberOfNodes),
		done:   make(chan struct{}),
	}, nil
}

func (c *EndCoordinatorController) Run(ctx context.Context) error {
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
					if c.state.WasSentGame(msg.GetClientID()) {
						slog.Info("Game END already propagated", "clientId", msg.GetClientID(), "node", msg.GetMessageID())
						delivery.Ack(false)
						continue
					}

					c.state.AddGame(msg.GetMessageID(), msg.GetClientID())
					c.state.PrintState()

					if c.state.AllGamesForId(msg.GetClientID()) {
						endMsg := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
							MessageID: msg.GetMessageID(),
							ClientID:  msg.GetClientID(),
							RequestID: msg.GetRequestID(),
						})
						<-time.After(3 * time.Second)
						slog.Info("Propagating END games", "counter", len(c.state.Games))
						if err := c.io.Write(endMsg.Marshal(), "game"); err != nil {
							return fmt.Errorf("couldn't write end message: %w", err)
						}

						// Reset state and add it to the sent clientIds
						c.state.ResetGame(msg.GetClientID())
						c.state.SetGameSent(msg.GetClientID())
						c.state.PrintState()
					}

					delivery.Ack(false)

				} else if msg.HasReviewData() {
					if c.state.WasSentReview(msg.GetClientID()) {
						slog.Info("Review END already propagated", "clientId", msg.GetClientID(), "node", msg.GetMessageID())
						delivery.Ack(false)
						continue
					}
					c.state.AddReview(msg.GetMessageID(), msg.GetClientID())
					c.state.PrintState()

					if c.state.AllReviewsForId(msg.GetClientID()) {
						endMsg := protocol.NewEndMessage(protocol.Reviews, protocol.MessageOptions{
							MessageID: msg.GetMessageID(),
							ClientID:  msg.GetClientID(),
							RequestID: msg.GetRequestID(),
						})
						<-time.After(8 * time.Second)
						slog.Info("Propagating END reviews", "counter", len(c.state.Reviews))
						if err := c.io.Write(endMsg.Marshal(), "review"); err != nil {
							return fmt.Errorf("couldn't write end message: %w", err)
						}

						// Reset state and add it to the sent clientIds
						c.state.ResetReview(msg.GetClientID())
						c.state.SetReviewSent(msg.GetClientID())
						c.state.PrintState()
					}

					delivery.Ack(false)
				} else {
					// Should never happen
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

func (c *EndCoordinatorController) Done() <-chan struct{} {
	return c.done
}

func (c *EndCoordinatorController) Close() {
	c.io.Close()
}
