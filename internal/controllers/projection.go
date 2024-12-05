package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/batch"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/end"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type Projection struct {
	iomanager client.IOManager
	done      chan struct{}
}

func NewProjection() (*Projection, error) {
	ioManager := client.IOManager{}
	err := ioManager.Connect(client.DirectSubscriber, client.Router)
	if err != nil {
		return nil, err
	}

	done := make(chan struct{}, 1)

	return &Projection{iomanager: ioManager, done: done}, nil
}

func (p *Projection) GetDone() <-chan struct{} {
	return p.done
}

func (p *Projection) DoneSignal() {
	p.done <- struct{}{}
}

func (p *Projection) Run(ctx context.Context) error {
	consumerChan := p.iomanager.Input.GetConsumer()
	defer p.DoneSignal()
	options, err := end.GetServiceOptionsFromEnv()
	if err != nil {
		return err
	}
	service, err := end.NewService(options)
	if err != nil {
		return err
	}
	ch := service.MergeConsumers(consumerChan)

	batcher := batch.NewBatcher(8500)
	for {
		select {
		case delivery := <-ch:
			var msg protocol.Message
			msgBytes := delivery.RecvDelivery.Body
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}
			if delivery.SenderType == end.SenderPrevious {
				batcher.Push(msg, delivery.RecvDelivery)
				if !batcher.IsFull() {
					continue
				}
				if err := p.processProjectionBatch(&batcher, service); err != nil {
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
			if err := p.processProjectionBatch(&batcher, service); err != nil {
				return fmt.Errorf("couldn't process batch: %w", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Projection) processProjectionBatch(batcher *batch.Batcher, endService *end.Service) error {
	currentBatch := batcher.Batch()
	for _, msg := range currentBatch {
		err := p.handleMessage(msg, endService)
		if err != nil {
			return err
		}
	}
	batcher.Acknowledge()
	return nil
}

// TODO(fede) - Replace name for something else
func (p *Projection) handleMessage(msg protocol.Message, service *end.Service) error {
	if msg.ExpectKind(protocol.Data) {
		if msg.HasGameData() {
			if err := p.handleGamesMessages(msg); err != nil {
				return err
			}
		} else if msg.HasReviewData() {
			if err := p.handleReviewsMessages(msg); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unexpected message that isn't games or reviews")
		}
	} else if msg.ExpectKind(protocol.End) {
		if msg.GetRequestID() == utils.MagicNumber {
			slog.Debug("End Msg with propagate number obtained", "clientId", msg.GetClientID(), "requestId", msg.GetRequestID())
			dataType := protocol.Reviews
			routerTag := "review"
			if msg.HasGameData() {
				dataType = protocol.Games
				routerTag = "game"
			}

			endMsg := protocol.NewEndMessage(dataType, protocol.MessageOptions{
				ClientID:  msg.GetClientID(),
				MessageID: msg.GetMessageID(),
				RequestID: 1,
			})

			slog.Debug("End Msg with propagate number sending", "router", routerTag, "clientId", msg.GetClientID(), "requestId", msg.GetRequestID())
			if err := p.iomanager.Write(endMsg.Marshal(), routerTag); err != nil {
				return fmt.Errorf("couldn't write end message: %w", err)
			}
		} else {
			slog.Debug("received end", "game", msg.HasGameData(), "reviews", msg.HasReviewData())
			service.NotifyNeighbours(msg)
		}
	} else {
		return fmt.Errorf("expected Data or End MessageType got %d", msg.GetMessageType())
	}
	return nil
}

// TODO(fede) - Replace hardcoded separators
func (p *Projection) handleGamesMessages(msg protocol.Message) error {
	elements := msg.Elements()
	var listOfGames []models.Game

	for _, element := range elements.Iter() {
		csvData := string(element.ReadBytes())
		reader := strings.NewReader(csvData)
		csvReader := csv.NewReader(reader)
		csvReader.LazyQuotes = true
		csvReader.FieldsPerRecord = -1

		listOfCsvGames, err := csvReader.ReadAll()
		if err != nil {
			return fmt.Errorf("could not parse lines of csv %s: %w", csvData, err)
		}

		for _, line := range listOfCsvGames {
			game, err := models.GameFromCSVLine(line)
			if err != nil {
				return fmt.Errorf("could not parse game from csv line %s: %w", line, err)
			}

			listOfGames = append(listOfGames, *game)
		}
	}

	for _, game := range listOfGames {
		payloadBuffer := protocol.NewPayloadBuffer(1)
		game.BuildPayload(payloadBuffer)

		responseMsg := protocol.NewDataMessage(protocol.Games, payloadBuffer.Bytes(), protocol.MessageOptions{
			MessageID: msg.GetMessageID(),
			ClientID:  msg.GetClientID(),
			RequestID: msg.GetRequestID(),
		})

		if err := p.iomanager.Write(responseMsg.Marshal(), "game"); err != nil {
			return err
		}
	}

	return nil
}

func (p *Projection) handleReviewsMessages(msg protocol.Message) error {
	elements := msg.Elements()
	var listOfReviews []models.Review

	for _, element := range elements.Iter() {
		csvData := string(element.ReadBytes())
		reader := strings.NewReader(csvData)
		csvReader := csv.NewReader(reader)
		csvReader.LazyQuotes = true
		csvReader.FieldsPerRecord = -1

		listOfCsvReviews, err := csvReader.ReadAll()
		if err != nil {
			return fmt.Errorf("could not parse lines of csv %s: %w", csvData, err)
		}

		for _, line := range listOfCsvReviews {
			review, err := models.ReviewFromCSVLine(line)
			if err != nil {
				return fmt.Errorf("could not parse review from csv line %s: %w", line, err)
			}

			listOfReviews = append(listOfReviews, *review)
		}
	}

	for _, review := range listOfReviews {
		payloadBuffer := protocol.NewPayloadBuffer(1)
		review.BuildPayload(payloadBuffer)

		responseMsg := protocol.NewDataMessage(protocol.Reviews, payloadBuffer.Bytes(), protocol.MessageOptions{
			MessageID: msg.GetMessageID(),
			ClientID:  msg.GetClientID(),
			RequestID: msg.GetRequestID(),
		})

		if err := p.iomanager.Write(responseMsg.Marshal(), "review"); err != nil {
			return err
		}
	}

	return nil
}

func (p *Projection) Close() {
	p.iomanager.Close()
}
