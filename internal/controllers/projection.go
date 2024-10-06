package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/rabbitmq/amqp091-go"
	"strings"
)

type Projection struct {
	iomanager client.IOManager
	done      chan struct{}
}

func NewProjection() (*Projection, error) {
	ioManager := client.IOManager{}
	err := ioManager.Connect(client.InputWorker, client.FanoutPublisher)
	if err != nil {
		return nil, err
	}

	done := make(chan struct{}, 1)

	return &Projection{ioManager, done}, nil
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

	for {
		select {
		case msg := <-consumerChan:
			response, err := p.handleMessage(msg)
			if err != nil {
				return err
			}

			// TODO(fede) - En el futuro se tiene que distinguir los ends y usar método de scincronización
			if response.HasGameData() {
				if err := p.iomanager.Write(response.Marshal(), "game"); err != nil {
					return err
				}
			} else if response.HasReviewData() {
				if err := p.iomanager.Write(response.Marshal(), "review"); err != nil {
					return err
				}
			}

			msg.Ack(false)
		case <-ctx.Done():
			return nil
		}
	}
}

// TODO(fede) - Replace name for something else
func (p *Projection) handleMessage(msg amqp091.Delivery) (*protocol.Message, error) {
	bytes := msg.Body
	internalMsg := protocol.Message{}
	err := internalMsg.Unmarshal(bytes)
	if err != nil {
		return nil, err
	}

	if internalMsg.ExpectKind(protocol.Data) {
		if internalMsg.HasGameData() {
			return p.handleGamesMessages(internalMsg)
		} else if internalMsg.HasReviewData() {
			return p.handleReviewsMessages(internalMsg)
		} else {
			return nil, fmt.Errorf("unexpected message that isn't games or reviews")
		}
	} else if internalMsg.ExpectKind(protocol.End) {
		return &internalMsg, nil
	} else {
		return nil, fmt.Errorf("expected Data or End MessageType got %d", internalMsg.GetMessageType())
	}
}

// TODO(fede) - Replace hardcoded separators
func (p *Projection) handleGamesMessages(msg protocol.Message) (*protocol.Message, error) {
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
			return nil, fmt.Errorf("could not parse lines of csv %s: %w", csvData, err)
		}

		for _, line := range listOfCsvGames {
			game, err := models.GameFromCSVLine(line)
			if err != nil {
				return nil, fmt.Errorf("could not parse game from csv line %s: %w", line, err)
			}

			listOfGames = append(listOfGames, *game)
		}
	}

	payloadBuffer := protocol.NewPayloadBuffer(len(listOfGames))
	for _, game := range listOfGames {
		game.BuildPayload(payloadBuffer)
	}

	responseMsg := protocol.NewDataMessage(protocol.Games, payloadBuffer.Bytes(), protocol.MessageOptions{
		MessageID: msg.GetMessageID(),
		ClientID:  msg.GetClientID(),
		RequestID: msg.GetRequestID(),
	})

	return &responseMsg, nil
}

func (p *Projection) handleReviewsMessages(msg protocol.Message) (*protocol.Message, error) {
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
			return nil, fmt.Errorf("could not parse lines of csv %s: %w", csvData, err)
		}

		for _, line := range listOfCsvReviews {
			review, err := models.ReviewFromCSVLine(line)
			if err != nil {
				return nil, fmt.Errorf("could not parse review from csv line %s: %w", line, err)
			}

			listOfReviews = append(listOfReviews, *review)
		}
	}

	payloadBuffer := protocol.NewPayloadBuffer(len(listOfReviews))
	for _, review := range listOfReviews {
		review.BuildPayload(payloadBuffer)
	}

	responseMsg := protocol.NewDataMessage(protocol.Reviews, payloadBuffer.Bytes(), protocol.MessageOptions{
		MessageID: msg.GetMessageID(),
		ClientID:  msg.GetClientID(),
		RequestID: msg.GetRequestID(),
	})

	return &responseMsg, nil
}
