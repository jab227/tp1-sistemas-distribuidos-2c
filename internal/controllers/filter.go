package controllers

import (
	"context"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/batch"
	"log/slog"
	"time"

	filter2 "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/filter"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/end"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"github.com/pemistahl/lingua-go"
)

var endFilterSupervisor = make(map[uint32]bool)

const MaxBatchingSize = 5000
const MaxBatcherTimeout = 10 * time.Second

type Filter struct {
	io client.IOManager

	gameFilter    filter2.FunFilterGames
	reviewFilter  filter2.FuncFilterReviews
	hasGameFilter bool

	detector *lingua.LanguageDetector

	done      chan struct{}
	clientID  uint32
	requestID uint32
}

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
	// Checks filter IO config
	filterIOConfig, ok := filter2.FilterInputsOutputs[filter]
	if !ok {
		return Filter{}, fmt.Errorf("unknown filter IO config: %s", filter)
	}
	slog.Debug("selected filter")
	if err := io.Connect(filterIOConfig.Input, filterIOConfig.Output); err != nil {
		return Filter{}, fmt.Errorf("couldn't create filter io: %w", err)
	}

	return Filter{
		io:            io,
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
	consumerCh := f.io.Input.GetConsumer()
	defer func() { f.done <- struct{}{} }()

	// Configure and starts End service
	options, err := end.GetServiceOptionsFromEnv()
	if err != nil {
		return err
	}
	service, err := end.NewService(options)
	ch := service.MergeConsumers(consumerCh)

	batcher := batch.NewBatcher(MaxBatchingSize)
	for {
		select {
		case delivery := <-ch:
			msgBytes := delivery.RecvDelivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}

			if delivery.SenderType == end.SenderPrevious {
				// Push to batcher and check if full
				batcher.Push(msg, delivery.RecvDelivery)
				if !batcher.IsFull() {
					continue
				}

				// Process batch
				if err := f.processBatch(&batcher, service); err != nil {
					return fmt.Errorf("couldn't process batch: %w", err)
				}
			} else if delivery.SenderType == end.SenderNeighbour {
				service.NotifyCoordinator(msg)
				delivery.RecvDelivery.Ack(false)
			} else {
				utils.Assert(false, "unknown type")
			}
		case <-time.After(MaxBatcherTimeout):
			if batcher.IsEmpty() {
				continue
			}

			// Process batch
			if err := f.processBatch(&batcher, service); err != nil {
				return fmt.Errorf("couldn't process batch: %w", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (f *Filter) processBatch(batch *batch.Batcher, endService *end.Service) error {
	currentBatch := batch.Batch()
	for _, msg := range currentBatch {
		if value, ok := endFilterSupervisor[msg.GetClientID()]; ok && value {
			slog.Debug("MSG AFTER END")
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
			endService.NotifyNeighbours(msg)
			endFilterSupervisor[msg.GetClientID()] = true
		}
	}

	batch.Acknowledge()
	return nil
}

func (f *Filter) handleGameFunc(receivedMsg protocol.Message) error {
	gamesPassed, err := f.gameFilter(receivedMsg)
	if err != nil {
		return fmt.Errorf("couldn't filter game: %w", err)
	}
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

		if err := f.io.Write(responseMsg.Marshal(), game.AppID); err != nil {
			return fmt.Errorf("couldn't write game response: %w", err)
		}
	}
	return nil
}

func (f *Filter) handleReviewFunc(receivedMsg protocol.Message) error {
	reviewsPassed, err := f.reviewFilter(receivedMsg, f.detector)
	if err != nil {
		return fmt.Errorf("couldn't filter reviews: %w", err)
	}
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

		if err := f.io.Write(responseMsg.Marshal(), review.AppID); err != nil {
			return fmt.Errorf("couldn't write review response: %w", err)
		}
	}
	return nil
}

func (f *Filter) Close() {
	f.io.Close()
}
