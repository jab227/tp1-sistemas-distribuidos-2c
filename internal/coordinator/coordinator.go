package coordinator

import (
	"context"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"log/slog"
)

type EndCoordinator struct {
	output client.OutputType
	io     client.IOManager

	expectedGamesEnd   int
	expectedReviewsEnd int

	GamesEndCounter   int
	ReviewsEndCounter int

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
		GamesEndCounter:    0,
		ReviewsEndCounter:  0,
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
					c.GamesEndCounter += 1
					slog.Info("Received Game END", "counter", c.GamesEndCounter)
					if c.GamesEndCounter >= c.expectedGamesEnd {
						endMsg := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
							MessageID: msg.GetMessageID(),
							ClientID:  msg.GetClientID(),
							RequestID: msg.GetRequestID(),
						})

						slog.Info("Propagating END", "counter", c.GamesEndCounter)
						// Hardcoded tag porque no nos importa a donde lo routee
						if err := c.io.Write(endMsg.Marshal(), "1"); err != nil {
							return fmt.Errorf("couldn't write end message: %w", err)
						}
					}

					// ACK of the MSg
					delivery.Ack(false)
				} else if msg.HasReviewData() {
					//. Check if its expects reviews
					if c.expectedReviewsEnd <= 0 {
						return fmt.Errorf("received reviews end but wasn't expected - Expected %d", c.expectedReviewsEnd)
					}

					// Sum counter
					c.ReviewsEndCounter += 1
					slog.Info("Received Review END", "counter", c.ReviewsEndCounter)
					if c.ReviewsEndCounter >= c.expectedReviewsEnd {
						endMsg := protocol.NewEndMessage(protocol.Reviews, protocol.MessageOptions{
							MessageID: msg.GetMessageID(),
							ClientID:  msg.GetClientID(),
							RequestID: msg.GetRequestID(),
						})

						slog.Info("Propagating END", "counter", c.ReviewsEndCounter)
						// Hardcoded tag porque no nos importa a donde lo routee
						if err := c.io.Write(endMsg.Marshal(), "1"); err != nil {
							return fmt.Errorf("couldn't write end message: %w", err)
						}
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
