package joinercoordinator

import (
	"context"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/controllers"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/persistence"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type EndJoinerCoordinatorController struct {
	output client.OutputType
	io     client.IOManager

	state         EndJoinerState
	numberOfNodes uint32
	propagateTags []string

	done chan struct{}
}

func NewEndJoinerCoordinatorController(output client.OutputType, numberOfNodes uint32, propagateTags []string) (*EndJoinerCoordinatorController, error) {
	io := client.IOManager{}
	if err := io.Connect(client.InputWorker, output); err != nil {
		return nil, fmt.Errorf("error initializing IOManager %s", err)
	}

	return &EndJoinerCoordinatorController{
		output:        output,
		io:            io,
		state:         *NewEndJoinerState(numberOfNodes),
		numberOfNodes: numberOfNodes,
		propagateTags: propagateTags,
		done:          make(chan struct{}),
	}, nil
}

func (c *EndJoinerCoordinatorController) Run(ctx context.Context, logFile string) error {
	endState, log, err := reloadEndJoinerState(logFile, c.numberOfNodes)
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
					if msg.GetRequestID() == utils.MagicNumber {
						if err := c.handleFinalReview(msg, delivery, log); err != nil {
							return fmt.Errorf("couldn't handle final review: %w", err)
						}
					} else {
						if err := c.handleReview(msg, delivery, log); err != nil {
							return fmt.Errorf("couldn't handle review message: %w", err)
						}
					}
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

func (c *EndJoinerCoordinatorController) handleReview(msg protocol.Message, delivery amqp091.Delivery, log *persistence.TransactionLog) error {
	if c.state.WasSentReview(msg.GetClientID()) {
		slog.Info("Review END already propagated", "clientId", msg.GetClientID(), "node", msg.GetMessageID())
		delivery.Ack(false)
		return nil
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
	return nil
}

func (c *EndJoinerCoordinatorController) handleFinalReview(msg protocol.Message, delivery amqp091.Delivery, log *persistence.TransactionLog) error {
	slog.Debug("Received final review", "clientId", msg.GetClientID(), "node", msg.GetMessageID())
	if c.state.WasSentFinalReview(msg.GetClientID()) {
		slog.Info("Final review END already propagated", "clientId", msg.GetClientID(), "node", msg.GetMessageID())
		delivery.Ack(false)
		return nil
	}
	c.state.AddFinalReview(msg.GetMessageID(), msg.GetClientID())
	c.state.PrintState()

	if c.state.AllFinalReviewsForId(msg.GetClientID()) {
		endMsg := protocol.NewEndMessage(protocol.Reviews, protocol.MessageOptions{
			MessageID: msg.GetMessageID(),
			ClientID:  msg.GetClientID(),
			RequestID: utils.MagicNumber2,
		})
		// <-time.After(8 * time.Second)

		for _, tag := range c.propagateTags {
			slog.Info("Propagating END reviews", "counter", len(c.state.Reviews), "requestId", endMsg.GetRequestID(), "tag", tag)
			if err := c.io.Write(endMsg.Marshal(), tag); err != nil {
				return fmt.Errorf("couldn't write end message: %w", err)
			}
		}

		// Reset state and add it to the sent clientIds
		c.state.ResetFinalReview(msg.GetClientID())
		c.state.SetFinalReview(msg.GetClientID())
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
	return nil

}

func (c *EndJoinerCoordinatorController) Done() <-chan struct{} {
	return c.done
}

func (c *EndJoinerCoordinatorController) Close() {
	c.io.Close()
}
