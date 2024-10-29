package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/heap"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

type topGamesState struct {
	heapGames *heap.HeapGames
}

type TopGames struct {
	iomanager client.IOManager
	done      chan struct{}
	state     *topGamesState
	n         uint64
}

func NewTopGames(n uint64) (*TopGames, error) {
	ioManager := client.IOManager{}
	err := ioManager.Connect(client.InputWorker, client.OutputWorker)
	if err != nil {
		return nil, err
	}

	return &TopGames{
		iomanager: ioManager,
		done:      make(chan struct{}, 1),
		state:     &topGamesState{heapGames: heap.NewHeapGames()},
		n:         n,
	}, nil
}

func (tg *TopGames) Done() <-chan struct{} {
	return tg.done
}

func (tg *TopGames) Run(ctx context.Context) error {
	consumerChan := tg.iomanager.Input.GetConsumer()
	defer func() {
		tg.done <- struct{}{}
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
				tg.processGamesData(internalMsg)
			} else if internalMsg.ExpectKind(protocol.End) {
				slog.Debug("received end", "game", internalMsg.HasGameData(), "review", internalMsg.HasReviewData())
				if err := tg.writeResult(internalMsg); err != nil {
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

func (tg *TopGames) processGamesData(internalMsg protocol.Message) {
	heapGames := tg.state.heapGames
	elements := internalMsg.Elements()

	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		if heapGames.Len() < int(tg.n) {
			heapGames.PushValue(game)
			continue
		}

		minGame := heapGames.PopValue()
		if game.AvgPlayTime < minGame.AvgPlayTime {
			heapGames.PushValue(minGame)
		} else {
			heapGames.PushValue(game)
		}
	}
}

func (tg *TopGames) writeResult(internalMsg protocol.Message) error {
	listOfGames := tg.state.heapGames.TopNGames(tg.n)
	slog.Debug("top10", "games", listOfGames)
	for _, game := range listOfGames {
		buffer := protocol.NewPayloadBuffer(1)
		buffer.BeginPayloadElement()
		buffer.WriteBytes([]byte(game.Name))
		buffer.EndPayloadElement()
		response := protocol.NewResultsMessage(protocol.Query2, buffer.Bytes(), protocol.MessageOptions{
			MessageID: internalMsg.GetMessageID(),
			ClientID:  internalMsg.GetClientID(),
			RequestID: internalMsg.GetRequestID(),
		})
		if err := tg.iomanager.Write(response.Marshal(), ""); err != nil {
			return fmt.Errorf("couldn't write query 2 output: %w", err)
		}
	}

	res := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
		MessageID: internalMsg.GetMessageID(),
		ClientID:  internalMsg.GetClientID(),
		RequestID: internalMsg.GetRequestID(),
	})
	// Tell it ends the query 2
	res.SetQueryResult(protocol.Query2)
	if err := tg.iomanager.Write(res.Marshal(), ""); err != nil {
		return fmt.Errorf("couldn't write query 2 end: %w", err)
	}
	slog.Debug("query 2 results", "state", listOfGames)

	// reset state
	tg.state.heapGames = heap.NewHeapGames()
	return nil
}
