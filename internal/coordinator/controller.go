package coordinator

import (
	"context"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/controllers"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"log/slog"
)

type EndCoordinatorController struct {
	output client.OutputType
	io     client.IOManager

	state         EndState
	numberOfNodes uint32
	propagateTags []string

	done chan struct{}
}

func NewEndCoordinatorController(output client.OutputType, numberOfNodes uint32, propagateTags []string) (*EndCoordinatorController, error) {
	io := client.IOManager{}
	if err := io.Connect(client.InputWorker, output); err != nil {
		return nil, fmt.Errorf("error initializing IOManager %s", err)
	}

	return &EndCoordinatorController{
		output:        output,
		io:            io,
		state:         *NewEndState(numberOfNodes),
		numberOfNodes: numberOfNodes,
		propagateTags: propagateTags,
		done:          make(chan struct{}),
	}, nil
}

func (c *EndCoordinatorController) Run(ctx context.Context, logFile string) error {
	endState, log, err := reloadEndState(logFile, c.numberOfNodes)
	if err != nil {
		return fmt.Errorf("error reloading EndState: %s", err)
	}

	c.state = endState

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
							RequestID: utils.MagicNumber,
						})
						// <-time.After(5 * time.Second)

						for _, tag := range c.propagateTags {
							slog.Info("Propagating END games", "counter", len(c.state.Games), "requestId", msg.GetRequestID(), "tag", tag)
							if err := c.io.Write(endMsg.Marshal(), tag); err != nil {
								return fmt.Errorf("couldn't write end message: %w", err)
							}
						}

						// Reset state and add it to the sent clientIds
						c.state.ResetGame(msg.GetClientID())
						c.state.SetGameSent(msg.GetClientID())
						c.state.PrintState()
					}

					stateBytes, err := MarshalState(&c.state)
					if err != nil {
						return fmt.Errorf("couldn't marshal state: %w", err)
					}
					log.Append(stateBytes, uint32(controllers.TXNSet))
					if err := log.Commit(); err != nil {
						return fmt.Errorf("couldn't commit log: %w", err)
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
							RequestID: utils.MagicNumber,
						})
						// <-time.After(8 * time.Second)

						for _, tag := range c.propagateTags {
							slog.Info("Propagating END reviews", "counter", len(c.state.Reviews), "requestId", endMsg.GetRequestID(), "tag", tag)
							if err := c.io.Write(endMsg.Marshal(), tag); err != nil {
								return fmt.Errorf("couldn't write end message: %w", err)
							}
						}

						// Reset state and add it to the sent clientIds
						c.state.ResetReview(msg.GetClientID())
						c.state.SetReviewSent(msg.GetClientID())
						c.state.PrintState()
					}

					stateBytes, err := MarshalState(&c.state)
					if err != nil {
						return fmt.Errorf("couldn't marshal state: %w", err)
					}
					log.Append(stateBytes, uint32(controllers.TXNSet))
					if err := log.Commit(); err != nil {
						return fmt.Errorf("couldn't commit log: %w", err)
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
