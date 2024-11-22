package controllers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/heap"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/batch"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
)

type topGamesState struct {
	heapGames *heap.HeapGames
}

type TopGames struct {
	iomanager client.IOManager
	done      chan struct{}
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

	topGamesStateStore := store.NewStore[*topGamesState]()
	batcher := batch.NewBatcher(50)
	for {
		select {
		case msg := <-consumerChan:
			bytes := msg.Body
			internalMsg := protocol.Message{}
			if err := internalMsg.Unmarshal(bytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}

			batcher.Push(internalMsg, msg)
			if !batcher.IsFull() {
				continue
			}
			batch := batcher.Batch()
			if err := processTopGamesBatch(batch, topGamesStateStore, tg); err != nil {
				return err
			}
			batcher.Acknowledge()
		case <-time.After(10 * time.Second):
			if batcher.IsEmpty() {
				continue
			}
			batch := batcher.Batch()
			if err := processTopGamesBatch(batch, topGamesStateStore, tg); err != nil {
				return err
			}
			batcher.Acknowledge()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func processTopGamesBatch(batch []protocol.Message, topGamesStateStore store.Store[*topGamesState], tg *TopGames) error {
	for _, internalMsg := range batch {
		clientID := internalMsg.GetClientID()
		state, ok := topGamesStateStore.Get(clientID)
		if !ok {
			state = &topGamesState{heapGames: heap.NewHeapGames()}
			topGamesStateStore.Set(clientID, state)
		}

		if internalMsg.ExpectKind(protocol.Data) {
			if !internalMsg.HasGameData() {
				return fmt.Errorf("wrong type: expected game data")
			}
			tg.processGamesData(state, internalMsg)
		} else if internalMsg.ExpectKind(protocol.End) {
			if err := tg.writeResult(state, internalMsg); err != nil {
				return err
			}
			topGamesStateStore.Delete(clientID)
		} else {
			return fmt.Errorf("unexpected message type: %s", internalMsg.GetMessageType())
		}
	}
	return nil
}

func (tg *TopGames) processGamesData(state *topGamesState, internalMsg protocol.Message) {
	heapGames := state.heapGames
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

func (tg *TopGames) writeResult(state *topGamesState, internalMsg protocol.Message) error {
	listOfGames := state.heapGames.TopNGames(tg.n)
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
	return nil
}
