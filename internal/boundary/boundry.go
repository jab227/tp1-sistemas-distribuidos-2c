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
	state  State

	senderCh chan protocol.Message
	errorsCh chan error

	ioManager *client.IOManager
}

func NewBoundary(config *Config, ioManager *client.IOManager) *Boundary {
	return &Boundary{
		config: *config,
		state:  *NewState(),

		senderCh:  make(chan protocol.Message),
		errorsCh:  make(chan error),
		ioManager: ioManager,
	}
}

// TODO(fede) - Handle signal
func (b *Boundary) Run() error {
	listenerSock, err := net.Listen("tcp", fmt.Sprintf("%s:%d", b.config.ServiceHost, b.config.ServicePort))
	if err != nil {
		return fmt.Errorf("could not listen to %s:%d: %v", b.config.ServiceHost, b.config.ServicePort, err)
	}

	go b.handleErrors()

	// Sender goroutine to handle message send
	go b.senderGoroutine()

	for {
		conn, err := listenerSock.Accept()
		if err != nil {
			return fmt.Errorf("could not accept connection: %v", err)
		}

		go b.handleClient(conn)
	}
}

func (b *Boundary) senderGoroutine() {
	for {
		select {
		case msg := <-b.senderCh:
			if err := b.ioManager.Write(msg.Marshal(), ""); err != nil {
				slog.Error("error with sender goroutine")
				b.errorsCh <- err
				return
			}
		}
	}
}

func (b *Boundary) handleErrors() {
	for {
		errMsg := <-b.errorsCh
		slog.Error(errMsg.Error())
	}
}

func (b *Boundary) handleClient(conn net.Conn) {
	defer conn.Close()

	resultsService := NewResultsService(conn, b.ioManager)
	if err := b.handleClientMsgs(conn); err != nil {
		b.errorsCh <- err
		return
	}

	// TODO(fede) - handle context
	if err := resultsService.Run(context.Background()); err != nil {
		slog.Error(err.Error())
		b.errorsCh <- err
		return
	}
}

func (b *Boundary) handleClientMsgs(conn net.Conn) error {
	endCounter := 0
	clientAddr := conn.RemoteAddr().String()

	clientId := b.state.GetNewClientId()
	requestId := b.state.GetClientNewRequestId(clientId)

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
			if err := b.handleRecvData(msg); err != nil {
				return fmt.Errorf("client: %s - %v", clientAddr, err)
			}

		} else if msg.IsEnd() {
			slog.Info("received end message",
				"clientId", msg.Header.ClientId,
				"requestId", msg.Header.RequestId,
				"messageId", b.state.GetClientNewMessageId(clientId),
			)

			if err := b.handleRecvEnd(msg); err != nil {
				return fmt.Errorf("client: %s - %v", clientAddr, err)
			}
			endCounter += 1
		}
	}

	return nil
}

func (b *Boundary) handleRecvData(msg cprotocol.Message) error {
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
			MessageID: b.state.GetClientNewMessageId(msg.Header.ClientId),
		},
	)

	b.senderCh <- internalMsg
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

	b.senderCh <- internalMsg
	return nil
}
