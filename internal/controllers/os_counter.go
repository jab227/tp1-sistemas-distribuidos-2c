package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/batch"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/persistence"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type newOSState map[string]models.OS

type osState struct {
	windows, mac, linux uint
}

func (o *osState) Update(state osState) {
	o.windows += state.windows
	o.mac += state.mac
	o.linux += state.linux
}

type OSCounter struct {
	io   client.IOManager
	done chan struct{}
}

func NewOSCounter() (OSCounter, error) {
	var io client.IOManager
	if err := io.Connect(client.DirectSubscriber, client.OutputWorker); err != nil {
		return OSCounter{}, fmt.Errorf("couldn't create os counter: %w", err)
	}
	return OSCounter{
		io:   io,
		done: make(chan struct{}, 1),
	}, nil
}

func (o OSCounter) Done() <-chan struct{} {
	return o.done
}

func applyOSBatch(
	batch []protocol.Message,
	osStateStore store.Store[newOSState],
) {
	for _, msg := range batch {
		if msg.ExpectKind(protocol.Data) {
			utils.Assert(msg.HasGameData(), "wrong type: expected game data")

			elements := msg.Elements()
			clientID := msg.GetClientID()
			state, ok := osStateStore.Get(clientID)
			if !ok {
				state = make(map[string]models.OS)
				osStateStore.Set(clientID, state)
			}
			for _, element := range elements.Iter() {
				game := models.ReadGame(&element)
				state[game.AppID] = game.SupportedOS
			}
		} else if msg.ExpectKind(protocol.End) {
			continue
		} else {
			utils.Assertf(false, "unexpected message type: %s", msg.GetMessageType())
		}
	}
}

func reloadOSCounter() (store.Store[newOSState], *persistence.TransactionLog, error) {
	stateStore := store.NewStore[newOSState]()
	log := persistence.NewTransactionLog("../logs/os_counter.log")
	logBytes, err := os.ReadFile("../logs/os_counter.log")
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
			applyOSBatch(currentBatch, stateStore)
		}
	}
	return stateStore, log, nil
}

func (o *OSCounter) Run(ctx context.Context) error {
	consumerCh := o.io.Input.GetConsumer()
	defer func() {
		o.done <- struct{}{}
	}()

	osStateStore, log, err := reloadOSCounter()
	if err != nil {
		return err
	}
	batcher := batch.NewBatcher(3000)

	for {
		select {
		case delivery := <-consumerCh:
			msgBytes := delivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}
			batcher.Push(msg, delivery)
			if !batcher.IsFull() {
				continue
			}
			currentBatch := batcher.Batch()
			log.Append(batch.MarshalBatch(currentBatch), uint32(TXNBatch))
			slog.Debug("processing batch")
			err := o.processBatchOsCounter(currentBatch, osStateStore)
			if err != nil {
				return err
			}
			slog.Debug("commit")
			if err := log.Commit(); err != nil {
				return fmt.Errorf("couldn't commit to disk: %w", err)
			}
			slog.Debug("acknowledge")
			batcher.Acknowledge()
		case <-time.After(10 * time.Second):
			if batcher.IsEmpty() {
				continue
			}
			currentBatch := batcher.Batch()
			log.Append(batch.MarshalBatch(currentBatch), uint32(TXNBatch))
			slog.Debug("processing batch")
			err := o.processBatchOsCounter(currentBatch, osStateStore)
			if err != nil {
				return err
			}
			slog.Debug("commit")
			if err := log.Commit(); err != nil {
				return fmt.Errorf("couldn't commit to disk: %w", err)
			}
			slog.Debug("acknowledge")
			batcher.Acknowledge()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *OSCounter) processBatchOsCounter(
	messages []protocol.Message,
	osStateStore store.Store[newOSState],
) error {
	for _, msg := range messages {
		if msg.ExpectKind(protocol.Data) {
			if !msg.HasGameData() {
				return fmt.Errorf("couldn't wrong type: expected game data")
			}

			elements := msg.Elements()
			clientID := msg.GetClientID()
			state, ok := osStateStore.Get(clientID)
			if !ok {
				state = make(map[string]models.OS)
				osStateStore.Set(clientID, state)
			}
			for _, element := range elements.Iter() {
				game := models.ReadGame(&element)
				state[game.AppID] = game.SupportedOS
			}
		} else if msg.ExpectKind(protocol.End) {
			clientID := msg.GetClientID()
			slog.Info("received end", "node", "os_counter")
			var s osState
			for _, o := range osStateStore[clientID] {
				if o.IsWindowsSupported() {
					s.windows += 1
				}
				if o.IsMacSupported() {
					s.mac += 1
				}
				if o.IsLinuxSupported() {
					s.linux += 1
				}
			}
			result := marshalResults(s)
			res := protocol.NewResultsMessage(protocol.Query1, result, protocol.MessageOptions{
				MessageID: msg.GetMessageID(),
				ClientID:  clientID,
				RequestID: msg.GetRequestID(),
			})
			if err := o.io.Write(res.Marshal(), ""); err != nil {
				return fmt.Errorf("couldn't write query 1 output: %w", err)
			}
			slog.Debug("query 1 results", "result", res, "state", osStateStore[clientID])
			osStateStore.Delete(clientID)
		} else {
			return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
		}
	}
	return nil
}

func marshalResults(s osState) []byte {
	builder := protocol.NewPayloadBuffer(1)
	builder.BeginPayloadElement()
	builder.WriteUint32(uint32(s.windows))
	builder.WriteUint32(uint32(s.mac))
	builder.WriteUint32(uint32(s.linux))
	builder.EndPayloadElement()
	return builder.Bytes()
}

func (o *OSCounter) Close() {
	o.io.Close()
}
