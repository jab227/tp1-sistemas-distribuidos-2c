package controllers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/heap"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

type topReviewsState struct {
	appByReviewScore map[string]int
}

type TopReviews struct {
	iomanager client.IOManager
	done      chan struct{}
	state     *topReviewsState
	n         uint64
}

func NewTopReviews(n uint64) (*TopReviews, error) {
	ioManager := client.IOManager{}
	err := ioManager.Connect(client.DirectSubscriber, client.OutputWorker)
	if err != nil {
		return nil, err
	}

	return &TopReviews{
		iomanager: ioManager,
		done:      make(chan struct{}, 1),
		state:     &topReviewsState{make(map[string]int)},
		n:         n,
	}, nil
}

func (tr *TopReviews) Done() <-chan struct{} {
	return tr.done
}

func (tr *TopReviews) Run(ctx context.Context) error {
	consumerChan := tr.iomanager.Input.GetConsumer()
	defer func() {
		tr.done <- struct{}{}
	}()

	for {
		select {
		case msg := <-consumerChan:
			bytes := msg.Body
			internalMsg := protocol.Message{}
			if err := internalMsg.Unmarshal(bytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}

			if internalMsg.ExpectKind(protocol.Data) {
				if !internalMsg.HasGameData() {
					return fmt.Errorf("wrong type: expected game data")
				}
				tr.processReviewsData(internalMsg)
			} else if internalMsg.ExpectKind(protocol.End) {
				if err := tr.writeResult(internalMsg); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unexpected message type: %s", internalMsg.GetMessageType())
			}

			msg.Ack(false)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (tr *TopReviews) processReviewsData(internalMsg protocol.Message) {
	elements := internalMsg.Elements()
	for _, element := range elements.Iter() {
		review := models.ReadReview(&element)
		key := fmt.Sprintf("%s,%s", review.AppID, review.Name)
		tr.state.appByReviewScore[key] += int(review.Score)
	}
}

func (tr *TopReviews) writeResult(internalMsg protocol.Message) error {
	reviews := make([]models.Review, 0, len(tr.state.appByReviewScore))
	for key, value := range tr.state.appByReviewScore {
		compositeKey := strings.Split(key, ",")
		review := models.Review{
			AppID: compositeKey[0],
			Name:  compositeKey[1],
			Score: models.ReviewScore(value),
		}
		reviews = append(reviews, review)
	}

	reviewsHeap := heap.NewHeapReviews()
	for _, review := range reviews {
		reviewsHeap.PushValue(review)
	}

	results := reviewsHeap.TopNReviews(int(tr.n))
	payloadBuffer := protocol.NewPayloadBuffer(len(results))
	for _, reviews := range results {
		reviews.BuildPayload(payloadBuffer)
	}

	response := protocol.NewResultsMessage(protocol.Query3, payloadBuffer.Bytes(), protocol.MessageOptions{
		MessageID: internalMsg.GetMessageID(),
		ClientID:  internalMsg.GetClientID(),
		RequestID: internalMsg.GetRequestID(),
	})

	if err := tr.iomanager.Write(response.Marshal(), ""); err != nil {
		return fmt.Errorf("couldn't write query 3 output: %w", err)
	}
	slog.Debug("query 3 results", "result", response, "state", *tr.state)

	// reset state
	tr.state.appByReviewScore = make(map[string]int)
	return nil
}
