package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

type osState struct {
	windows, mac, linux uint
}

type OSCounter struct {
	io   client.IOManager
	done chan struct{}
	s    osState
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
				for _, element := range elements.Iter() {
					game := models.ReadGame(&element)
					if game.SupportedOS.IsWindowsSupported() {
						o.s.windows += 1
					}
					if game.SupportedOS.IsMacSupported() {
						o.s.mac += 1
					}
					if game.SupportedOS.IsLinuxSupported() {
						o.s.linux += 1
					}
				}
			} else if msg.ExpectKind(protocol.End) {
				// reset state
				slog.Debug("received end", "node", "os_counter")
				builder := protocol.NewPayloadBuffer(1)
				builder.BeginPayloadElement()
				builder.WriteUint32(uint32(o.s.windows))
				builder.WriteUint32(uint32(o.s.mac))
				builder.WriteUint32(uint32(o.s.linux))
				builder.EndPayloadElement()
				res := protocol.NewResultsMessage(protocol.Query1, builder.Bytes(), protocol.MessageOptions{
					MessageID: msg.GetMessageID(),
					ClientID:  msg.GetClientID(),
					RequestID: msg.GetRequestID(),
				})
				if err := o.io.Write(res.Marshal(), ""); err != nil {
					return fmt.Errorf("couldn't write query 1 output: %w", err)
				}
				slog.Debug("query 1 results", "result", res, "state", o.s)
				o.s = osState{}
			} else {
				return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
			}
			delivery.Ack(false)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *OSCounter) Close() {
	o.io.Close()
}
