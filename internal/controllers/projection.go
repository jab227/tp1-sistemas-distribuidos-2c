package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/end"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"github.com/rabbitmq/amqp091-go"
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
	for {
		select {
		case delivery := <-ch:
			if delivery.SenderType == end.SenderPrevious {
				err := p.handleMessage(delivery.RecvDelivery, service)
				if err != nil {
					return err
				}
				delivery.RecvDelivery.Ack(false)
			} else if delivery.SenderType == end.SenderNeighbour {
				msgBytes := delivery.RecvDelivery.Body
				var msg protocol.Message
				if err := msg.Unmarshal(msgBytes); err != nil {
					slog.Error("couldn't unmarshal message", "error", err)
					continue
				}
				service.NotifyCoordinator(msg)
				delivery.RecvDelivery.Ack(false)
			} else {
				utils.Assert(false, "unknown type")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// TODO(fede) - Replace name for something else
func (p *Projection) handleMessage(msg amqp091.Delivery, service *end.Service) error {
	bytes := msg.Body
	internalMsg := protocol.Message{}
	err := internalMsg.Unmarshal(bytes)
	if err != nil {
		return err
	}
	if internalMsg.ExpectKind(protocol.Data) {
		if internalMsg.HasGameData() {
			if err = p.handleGamesMessages(internalMsg); err != nil {
				return err
			}
		} else if internalMsg.HasReviewData() {
			if err = p.handleReviewsMessages(internalMsg); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unexpected message that isn't games or reviews")
		}
	} else if internalMsg.ExpectKind(protocol.End) {
		slog.Debug("received end", "game", internalMsg.HasGameData(), "reviews", internalMsg.HasReviewData())
		service.NotifyNeighbours(internalMsg)
	} else {
		return fmt.Errorf("expected Data or End MessageType got %d", internalMsg.GetMessageType())
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
