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
	n         int
}

func NewTopReviews(n int) (*TopReviews, error) {
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
	consumerCh := tr.iomanager.Input.GetConsumer()
	defer func() {
		tr.done <- struct{}{}
	}()

	for {
		select {
		case delivery := <-consumerCh:
			bytes := delivery.Body
			msg := protocol.Message{}
			if err := msg.Unmarshal(bytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}

			if msg.ExpectKind(protocol.Data) {
				if !msg.HasGameData() {
					return fmt.Errorf("wrong type: expected game data")
				}
				tr.processReviewsData(msg)
			} else if msg.ExpectKind(protocol.End) {
				slog.Debug("received end", "game", msg.HasGameData())
				if err := tr.writeResult(msg); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
			}
			delivery.Ack(false)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (tr *TopReviews) processReviewsData(internalMsg protocol.Message) {
	elements := internalMsg.Elements()
	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		slog.Debug("received game", "game", game)
		key := fmt.Sprintf("%s||%s", game.AppID, game.Name)
		tr.state.appByReviewScore[key] += 1
	}
}

func (tr *TopReviews) writeResult(internalMsg protocol.Message) error {
	values := make([]heap.Value, 0, len(tr.state.appByReviewScore))
	for k, v := range tr.state.appByReviewScore {
		name := strings.Split(k, "||")[1]
		count := v
		values = append(values, heap.Value{name, count})
	}
	h := heap.NewHeap()
	for _, value := range values {
		h.PushValue(value)
	}

	results := h.TopN(int(tr.n))
	slog.Debug("topn results", "results", results)
	buffer := protocol.NewPayloadBuffer(len(results))
	for _, value := range results {
		buffer.BeginPayloadElement()
		buffer.WriteBytes([]byte(value.Name))
		buffer.EndPayloadElement()
	}

	response := protocol.NewResultsMessage(protocol.Query3, buffer.Bytes(), protocol.MessageOptions{
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

func (t *TopReviews) Close() {
	t.iomanager.Close()
}
