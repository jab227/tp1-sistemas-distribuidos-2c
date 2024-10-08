package controllers

import (
	"context"
	"fmt"
	"log/slog"

	filter2 "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/filter"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/routing"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/pemistahl/lingua-go"
	"github.com/rabbitmq/amqp091-go"
)

type Filter struct {
	io         client.IOManager
	router     routing.Router
	usesRouter bool

	gameFilter    filter2.FunFilterGames
	reviewFilter  filter2.FuncFilterReviews
	hasGameFilter bool

	detector *lingua.LanguageDetector

	done chan struct{}
}

// TODO - Improve this brainrot code
// GRUG APPROVED
func NewFilter(filter string) (Filter, error) {
	// Check if it is a game filter
	var gameFilterFunc filter2.FunFilterGames
	var reviewFilterFunc filter2.FuncFilterReviews
	var hasGameFilter bool
	var ok bool
	slog.Debug("Entered NewFilter")
	gameFilterFunc, ok = filter2.FilterGamesMap[filter]
	if !ok {
		reviewFilterFunc, ok = filter2.FilterReviewsMap[filter]
		if !ok {
			return Filter{}, fmt.Errorf("unknown filter: %s", filter)
		}
		hasGameFilter = false
	} else {
		hasGameFilter = true
	}

	// If it needs a detector, create one
	var detector *lingua.LanguageDetector
	if filter2.NeedsDecoder(filter) {
		languages := []lingua.Language{
			lingua.English,
			lingua.Spanish,
		}
		detectorObj := lingua.NewLanguageDetectorBuilder().FromLanguages(languages...).Build()
		detector = &detectorObj
	}

	var io client.IOManager
	var router routing.Router
	// Checks filter IO config
	filterIOConfig, ok := filter2.FilterInputsOutputs[filter]
	if !ok {
		return Filter{}, fmt.Errorf("unknown filter IO config: %s", filter)
	}
	slog.Debug("selected filter")
	if filterIOConfig.UseRouter {
		slog.Debug("use router")
		internalRouter, err := routing.NewRouterById(filterIOConfig.Input)
		if err != nil {
			return Filter{}, fmt.Errorf("failed to create filter router: %w", err)
		}
		router = internalRouter
	} else {
		slog.Debug("dont use router")
		if err := io.Connect(filterIOConfig.Input, filterIOConfig.Output); err != nil {
			return Filter{}, fmt.Errorf("couldn't create filter io: %w", err)
		}
	}

	return Filter{
		io:         io,
		router:     router,
		usesRouter: filterIOConfig.UseRouter,

		gameFilter:    gameFilterFunc,
		reviewFilter:  reviewFilterFunc,
		hasGameFilter: hasGameFilter,
		detector:      detector,

		done: make(chan struct{}),
	}, nil
}

func (f *Filter) Done() <-chan struct{} {
	return f.done
}

func (f *Filter) Run(ctx context.Context) error {
	var consumerCh <-chan amqp091.Delivery
	if !f.usesRouter {
		consumerCh = f.io.Input.GetConsumer()
	} else {
		consumerCh = f.router.GetConsumer()
	}
	defer func() { f.done <- struct{}{} }()

	for {
		select {
		case delivery := <-consumerCh:
			msgBytes := delivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}

			// Detect type
			if msg.ExpectKind(protocol.Data) {
				// Handle filter
				if f.hasGameFilter {
					if err := f.handleGameFunc(msg); err != nil {
						return fmt.Errorf("couldn't handle game function: %w", err)
					}
				} else {
					if err := f.handleReviewFunc(msg); err != nil {
						return fmt.Errorf("couldn't handle review function: %w", err)
					}
				}
			} else if msg.ExpectKind(protocol.End) {
				var msgType protocol.DataType
				if msg.HasGameData() {
					msgType = protocol.Games
				} else if msg.HasReviewData() {
					msgType = protocol.Reviews
				}

				newMsg := protocol.NewEndMessage(
					msgType,
					protocol.MessageOptions{
						ClientID:  msg.GetClientID(),
						RequestID: msg.GetRequestID(),
						MessageID: msg.GetMessageID(),
					})
				if !f.usesRouter {
					if err := f.io.Write(newMsg.Marshal(), ""); err != nil {
						return fmt.Errorf("couldn't write end message: %w", err)
					}
				} else {
					if err := f.router.Write(newMsg.Marshal(), ""); err != nil {
						return fmt.Errorf("couldn't write end message: %w", err)
					}
				}
				slog.Info("Received End message",
					"clientId", msg.GetClientID(),
					"requestId", msg.GetRequestID(),
					"type", msg.GetRequestID(),
				)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (f *Filter) handleGameFunc(receivedMsg protocol.Message) error {
	gamesPassed, err := f.gameFilter(receivedMsg)
	if err != nil {
		return fmt.Errorf("couldn't filter game: %w", err)
	}
	slog.Debug("games passed", "games", gamesPassed)
	if f.usesRouter {
		for _, game := range gamesPassed {
			payloadBuffer := protocol.NewPayloadBuffer(1)
			payloadBuffer.BeginPayloadElement()
			game.BuildPayload(payloadBuffer)
			payloadBuffer.EndPayloadElement()

			responseMsg := protocol.NewDataMessage(
				protocol.Games,
				payloadBuffer.Bytes(),
				protocol.MessageOptions{
					ClientID:  receivedMsg.GetClientID(),
					RequestID: receivedMsg.GetRequestID(),
					MessageID: receivedMsg.GetMessageID(),
				},
			)

			if err := f.router.Write(responseMsg.Marshal(), game.AppID); err != nil {
				return fmt.Errorf("couldn't write game response: %w", err)
			}
		}
	} else {
		payloadBuffer := protocol.NewPayloadBuffer(len(gamesPassed))
		for _, game := range gamesPassed {
			payloadBuffer.BeginPayloadElement()
			game.BuildPayload(payloadBuffer)
			payloadBuffer.EndPayloadElement()
		}

		responseMsg := protocol.NewDataMessage(
			protocol.Games,
			payloadBuffer.Bytes(),
			protocol.MessageOptions{
				ClientID:  receivedMsg.GetClientID(),
				RequestID: receivedMsg.GetRequestID(),
				MessageID: receivedMsg.GetMessageID(),
			},
		)

		if err := f.io.Write(responseMsg.Marshal(), ""); err != nil {
			return fmt.Errorf("couldn't write game message: %w", err)
		}
	}

	return nil
}

func (f *Filter) handleReviewFunc(receivedMsg protocol.Message) error {
	reviewsPassed, err := f.reviewFilter(receivedMsg, f.detector)
	if err != nil {
		return fmt.Errorf("couldn't filter reviews: %w", err)
	}

	if f.usesRouter {
		for _, review := range reviewsPassed {
			payloadBuffer := protocol.NewPayloadBuffer(1)
			payloadBuffer.BeginPayloadElement()
			review.BuildPayload(payloadBuffer)
			payloadBuffer.EndPayloadElement()

			responseMsg := protocol.NewDataMessage(
				protocol.Reviews,
				payloadBuffer.Bytes(),
				protocol.MessageOptions{
					ClientID:  receivedMsg.GetClientID(),
					RequestID: receivedMsg.GetRequestID(),
					MessageID: receivedMsg.GetMessageID(),
				},
			)

			if err := f.router.Write(responseMsg.Marshal(), review.AppID); err != nil {
				return fmt.Errorf("couldn't write review response: %w", err)
			}
		}
	} else {
		payloadBuffer := protocol.NewPayloadBuffer(len(reviewsPassed))
		for _, review := range reviewsPassed {
			payloadBuffer.BeginPayloadElement()
			review.BuildPayload(payloadBuffer)
			payloadBuffer.EndPayloadElement()
		}

		responseMsg := protocol.NewDataMessage(
			protocol.Reviews,
			payloadBuffer.Bytes(),
			protocol.MessageOptions{
				ClientID:  receivedMsg.GetClientID(),
				RequestID: receivedMsg.GetRequestID(),
				MessageID: receivedMsg.GetMessageID(),
			},
		)

		if err := f.io.Write(responseMsg.Marshal(), ""); err != nil {
			return fmt.Errorf("couldn't write review message: %w", err)
		}
	}

	return nil
}

func (f *Filter) Close() {
	if f.usesRouter {
		f.router.Destroy()
	} else {
		f.io.Close()
	}
}
