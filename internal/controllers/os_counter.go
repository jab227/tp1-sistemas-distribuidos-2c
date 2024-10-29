package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/store"
)

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

func (o *OSCounter) Run(ctx context.Context) error {
	consumerCh := o.io.Input.GetConsumer()
	defer func() {
		o.done <- struct{}{}
	}()
	osStateStore := store.NewStore[osState]()
	for {
		select {
		case delivery := <-consumerCh:
			msgBytes := delivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}
			if msg.ExpectKind(protocol.Data) {
				if !msg.HasGameData() {
					return fmt.Errorf("couldn't wrong type: expected game data")
				}

				elements := msg.Elements()
				clientID := msg.GetClientID()
				var s osState
				for _, element := range elements.Iter() {
					game := models.ReadGame(&element)
					if game.SupportedOS.IsWindowsSupported() {
						s.windows += 1
					}
					if game.SupportedOS.IsMacSupported() {
						s.mac += 1
					}
					if game.SupportedOS.IsLinuxSupported() {
						s.linux += 1
					}
				}

				state, ok := osStateStore.Get(clientID)
				if ok {
					s.Update(state)
				}
				osStateStore.Set(clientID, s)
			} else if msg.ExpectKind(protocol.End) {
				// reset state
				clientID := msg.GetClientID()
				slog.Info("received end", "node", "os_counter")
				result := marshalResults(osStateStore[clientID])
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
			delivery.Ack(false)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
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
