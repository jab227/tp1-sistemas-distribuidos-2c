package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/heap"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/batch"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/persistence"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type topGamesState struct {
	// heapGames *heap.HeapGames
	appByGameScore map[string]int
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

	topGamesStateStore, log, err := reloadTopGames(tg)
	if err != nil {
		return err
	}
	batcher := batch.NewBatcher(3000)

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
			currentBatch := batcher.Batch()
			log.Append(batch.MarshalBatch(currentBatch), uint32(TXNBatch))

			if err := processTopGamesBatch(currentBatch, topGamesStateStore, tg); err != nil {
				return err
			}

			if err := log.Commit(); err != nil {
				return fmt.Errorf("couldn't commit to disk: %w", err)
			}
			time.Sleep(5 * time.Second)

			batcher.Acknowledge()
		case <-time.After(10 * time.Second):
			if batcher.IsEmpty() {
				continue
			}
			currentBatch := batcher.Batch()
			log.Append(batch.MarshalBatch(currentBatch), uint32(TXNBatch))

			if err := processTopGamesBatch(currentBatch, topGamesStateStore, tg); err != nil {
				return err
			}

			if err := log.Commit(); err != nil {
				return fmt.Errorf("couldn't commit to disk: %w", err)
			}
			time.Sleep(5 * time.Second)

			batcher.Acknowledge()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func reloadTopGames(tg *TopGames) (store.Store[*topGamesState], *persistence.TransactionLog, error) {
	stateStore := store.NewStore[*topGamesState]()
	log := persistence.NewTransactionLog("../logs/top_games.log")
	logBytes, err := os.ReadFile("../logs/top_games.log")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return stateStore, log, nil
		}
		return stateStore, log, err
	}
	if err := log.Unmarshal(logBytes); err != nil {
		err = fmt.Errorf("couldn't unmarshal log: %w", err)
		return stateStore, log, err
	}
	for _, entry := range log.GetLog() {
		switch TXN(entry.Kind) {
		case TXNBatch:
			currentBatch, err := batch.UnmarshalBatch(entry.Data)
			if err != nil {
				return stateStore, log, err
			}
			applyTopGamesBatch(currentBatch, stateStore, tg)
		}
	}
	return stateStore, log, nil
}

func applyTopGamesBatch(
	batch []protocol.Message,
	topGamesStateStore store.Store[*topGamesState],
	tg *TopGames,
) {
	for _, msg := range batch {
		if msg.ExpectKind(protocol.Data) {
			utils.Assert(msg.HasGameData(), "wrong type: expected game data")
			clientID := msg.GetClientID()
			state, ok := topGamesStateStore.Get(clientID)
			if !ok {
				state = &topGamesState{make(map[string]int)}
				topGamesStateStore.Set(clientID, state)
			}
			tg.processGamesData(state, msg)
		} else if msg.ExpectKind(protocol.End) {
			continue
		} else {
			utils.Assertf(false, "unexpected message type: %s", msg.GetMessageType())
		}
	}
}

func processTopGamesBatch(batch []protocol.Message, topGamesStateStore store.Store[*topGamesState], tg *TopGames) error {
	for _, internalMsg := range batch {
		clientID := internalMsg.GetClientID()
		state, ok := topGamesStateStore.Get(clientID)
		if !ok {
			state = &topGamesState{make(map[string]int)}
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
	elements := internalMsg.Elements()
	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		key := fmt.Sprintf("%s||%s", game.AppID, game.Name)
		state.appByGameScore[key] = int(game.AvgPlayTime)
	}
}

func (tg *TopGames) writeResult(state *topGamesState, internalMsg protocol.Message) error {
	values := make([]heap.Value, 0, len(state.appByGameScore))
	for k, v := range state.appByGameScore {
		name := strings.Split(k, "||")[1]
		avgPlayTime := v
		values = append(values, heap.Value{Name: name, Count: avgPlayTime})
	}
	h := heap.NewHeap()
	for _, value := range values {
		h.PushValue(value)
	}

	listOfGames := h.TopN(int(tg.n))
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
