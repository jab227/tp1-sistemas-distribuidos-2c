package src

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/message"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/utils"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/network"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"log/slog"
)

type Client struct {
	socket       *network.SocketTcp
	deleteSocket func()
	protocol     *communication.Protocol

	msgIdCounter uint32
	clientId     uint32
}

func NewClient(socket *network.SocketTcp, clientId uint32, deleteSocket func()) (*Client, func(), error) {
	client := &Client{
		socket:       socket,
		deleteSocket: deleteSocket,
		msgIdCounter: 0,
		clientId:     clientId,
	}
	client.protocol = communication.NewProtocol(socket)
	if err := client.protocol.SyncAck(clientId); err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		deleteClient(client)
	}
	return client, cleanup, nil
}

func deleteClient(c *Client) {
	c.deleteSocket()
}

func (c *Client) GetMessageId() uint32 {
	defer func() { c.msgIdCounter++ }()
	return c.msgIdCounter
}

// TODO - Crear abstraccion para validaciones
// TODO(fede) - Parte 2 - Validar esta comunicación con el IOManager es thread safe
func (c *Client) Execute(ioManager *client.IOManager) error {
	for {
		msgData, err := c.protocol.RecvDataMessage()
		if err != nil {
			return err
		}

		if msgData.Header.Optype == message.Data {
			// This payload contains data
			if msgData.Payload.Header.Start == utils.StartNotSet && msgData.Payload.Header.End != utils.EndNotSet {
				if msgData.Payload.Header.Type == uint8(message.Games) {
					internalMsg := protocol.NewDataMessage(protocol.Games,
						msgData.Payload.Payload.Marshall(),
						protocol.MessageOptions{
							ClientID:  msgData.Header.ClientId,
							RequestID: msgData.Header.RequestId,
							MessageID: c.GetMessageId(),
						},
					)

					err = ioManager.Write(internalMsg.Marshal(), "")
					if err != nil {
						return fmt.Errorf("cannot send data to client: %w - %v", err, internalMsg)
					}
				} else if msgData.Payload.Header.Type == uint8(message.Reviews) {
					internalMsg := protocol.NewDataMessage(protocol.Reviews,
						msgData.Payload.Payload.Marshall(),
						protocol.MessageOptions{
							ClientID:  msgData.Header.ClientId,
							RequestID: msgData.Header.RequestId,
							MessageID: c.GetMessageId(),
						},
					)

					err = ioManager.Write(internalMsg.Marshal(), "")
					if err != nil {
						return fmt.Errorf("cannot send data to client: %w - %v", err, internalMsg)
					}
				}
			} else if msgData.Payload.Header.Start == utils.StartSet && msgData.Payload.Header.End == utils.EndNotSet {
				// Handle start
				slog.Info("Received start message",
					"clientId", msgData.Header.ClientId,
					"requestId", msgData.Header.RequestId,
					"type", msgData.Payload.Header.Type,
				)
			} else if msgData.Payload.Header.End == utils.EndSet && msgData.Payload.Header.Start == utils.StartNotSet {
				if msgData.Payload.Header.Type == uint8(message.Games) {
					// Handle games
					internalMsg := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
						ClientID:  msgData.Header.ClientId,
						RequestID: msgData.Header.RequestId,
						MessageID: 1, // TODO(fede) - Check if matters
					})

					err = ioManager.Write(internalMsg.Marshal(), "")
					if err != nil {
						return fmt.Errorf("cannot send data to client: %w - %v", err, internalMsg)
					}
					slog.Info("Received End message",
						"clientId", msgData.Header.ClientId,
						"requestId", msgData.Header.RequestId,
						"type", msgData.Payload.Header.Type,
					)
				} else if msgData.Payload.Header.Type == uint8(message.Reviews) {
					// Handle reviews
					internalMsg := protocol.NewEndMessage(protocol.Reviews, protocol.MessageOptions{
						ClientID:  msgData.Header.ClientId,
						RequestID: msgData.Header.RequestId,
						MessageID: 1, // TODO(fede) - Check if matters
					})

					err = ioManager.Write(internalMsg.Marshal(), "")
					if err != nil {
						return fmt.Errorf("cannot send data to client: %w - %v", err, internalMsg)
					}

					slog.Info("Received End message",
						"clientId", msgData.Header.ClientId,
						"requestId", msgData.Header.RequestId,
						"type", msgData.Payload.Header.Type,
					)
				}
			}
		}
	}
}
