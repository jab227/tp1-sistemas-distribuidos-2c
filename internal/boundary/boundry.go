package boundary

import (
	"context"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/cprotocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"log/slog"
	"net"
)

type Boundary struct {
	config Config

	// TODO(fede) - Add sync for multiple processes accessing to this
	ioManager *client.IOManager
}

func NewBoundary(config *Config, ioManager *client.IOManager) *Boundary {
	return &Boundary{
		config:    *config,
		ioManager: ioManager,
	}
}

// TODO(fede) - Handle signal
func (b *Boundary) Run() error {
	listenerSock, err := net.Listen("tcp", fmt.Sprintf("%s:%d", b.config.ServiceHost, b.config.ServicePort))
	if err != nil {
		return fmt.Errorf("could not listen to %s:%d: %v", b.config.ServiceHost, b.config.ServicePort, err)
	}

	// Error goroutine to handle errors
	errorsCh := make(chan error)
	go b.handleErrors(errorsCh)

	for {
		conn, err := listenerSock.Accept()
		if err != nil {
			return fmt.Errorf("could not accept connection: %v", err)
		}

		go b.handleClient(conn, errorsCh)
	}
}

func (b *Boundary) handleErrors(errorsCh <-chan error) {
	for {
		errMsg := <-errorsCh
		slog.Error(errMsg.Error())
	}
}

func (b *Boundary) handleClient(conn net.Conn, errorsCh chan<- error) {
	defer conn.Close()

	resultsService := NewResultsService(conn, b.ioManager)
	if err := b.handleClientMsgs(conn); err != nil {
		errorsCh <- err
		return
	}

	// TODO(fede) - handle context
	if err := resultsService.Run(context.Background()); err != nil {
		slog.Error(err.Error())
		errorsCh <- err
		return
	}
}

func (b *Boundary) handleClientMsgs(conn net.Conn) error {
	endCounter := 0
	clientAddr := conn.RemoteAddr().String()

	// TODO(fede) - Replace hardcoded Ids
	clientId := uint64(1)
	requestId := uint64(1)
	messageId := uint32(1)

	// Wait for SyncMsg
	_, err := cprotocol.ReadSyncMsg(conn)
	if err != nil {
		return fmt.Errorf("client: %s - could not read sync message: %v", clientAddr, err)
	}

	// Response with AckSync
	if err := cprotocol.SendAckSyncMsg(conn, clientId, requestId); err != nil {
		return fmt.Errorf("client: %s - could not send ack: %v", clientAddr, err)
	}

	for {
		if endCounter == 2 {
			slog.Info("received all ends")
			break
		}

		msg, err := cprotocol.ReadMsg(conn)
		if err != nil {
			return fmt.Errorf("client: %s - could not read message: %v", clientAddr, err)
		}

		if msg.IsStart() {
			slog.Info("received start message",
				"type", msg.Header.ContentType,
				"requestId", msg.Header.RequestId,
				"clientId", msg.Header.ClientId,
			)
		} else if msg.IsData() {
			if err := b.handleRecvData(msg, messageId); err != nil {
				return fmt.Errorf("client: %s - %v", clientAddr, err)
			}

		} else if msg.IsEnd() {
			slog.Info("received end message",
				"clientId", msg.Header.ClientId,
				"requestId", msg.Header.RequestId,
				"messageId", messageId,
			)

			if err := b.handleRecvEnd(msg); err != nil {
				return fmt.Errorf("client: %s - %v", clientAddr, err)
			}
			endCounter += 1
		}
	}

	return nil
}

// TODO(fede) - Better handle messageId
func (b *Boundary) handleRecvData(msg cprotocol.Message, messageId uint32) error {
	payloadBuffer := protocol.NewPayloadBuffer(1)
	payloadBuffer.BeginPayloadElement()
	payloadBuffer.WriteBytes(msg.Payload)
	payloadBuffer.EndPayloadElement()

	msgDataType := protocol.Games
	if msg.IsGamesMsg() {
		msgDataType = protocol.Games
	} else if msg.IsReviewsMsg() {
		msgDataType = protocol.Reviews
	}

	internalMsg := protocol.NewDataMessage(msgDataType,
		payloadBuffer.Bytes(),
		protocol.MessageOptions{
			ClientID:  uint32(msg.Header.ClientId),
			RequestID: uint32(msg.Header.RequestId),
			MessageID: messageId,
		},
	)

	if err := b.ioManager.Write(internalMsg.Marshal(), ""); err != nil {
		return fmt.Errorf("cannot send data to projection: %w - %v", err, internalMsg)
	}

	return nil
}

func (b *Boundary) handleRecvEnd(msg cprotocol.Message) error {
	msgDataType := protocol.Games
	if msg.IsGamesMsg() {
		msgDataType = protocol.Games
	} else if msg.IsReviewsMsg() {
		msgDataType = protocol.Reviews
	}

	internalMsg := protocol.NewEndMessage(msgDataType, protocol.MessageOptions{
		MessageID: 1, //TODO(fede) - Check if it changes something
		ClientID:  uint32(msg.Header.ClientId),
		RequestID: uint32(msg.Header.RequestId),
	})

	if err := b.ioManager.Write(internalMsg.Marshal(), ""); err != nil {
		return fmt.Errorf("cannot send END msg to projection: %w", err)
	}

	return nil
}
