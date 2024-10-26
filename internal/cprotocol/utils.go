package cprotocol

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"net"
)

func SendSyncMsg(conn net.Conn) error {
	msg := NewSyncMessage()
	if err := utils.WriteToSocket(conn, msg.Marshal()); err != nil {
		return err
	}

	return nil
}

func SendAckSyncMsg(conn net.Conn, clientId uint64, requestId uint64) error {
	msg := NewAckSyncMessage(clientId, requestId)
	if err := utils.WriteToSocket(conn, msg.Marshal()); err != nil {
		return err
	}

	return nil
}

func SendStartMsg(conn net.Conn, fileType FileType, clientId uint64, requestId uint64) error {
	msg := NewStartMessage(fileType, clientId, requestId)
	if err := utils.WriteToSocket(conn, msg.Marshal()); err != nil {
		return err
	}

	return nil
}

func SendEndMsg(conn net.Conn, fileType FileType, clientId uint64, requestId uint64) error {
	msg := NewEndMessage(fileType, clientId, requestId)
	if err := utils.WriteToSocket(conn, msg.Marshal()); err != nil {
		return err
	}

	return nil
}

func SendDataMsg(conn net.Conn, fileType FileType, clientId uint64, requestId uint64, data []byte) error {
	msg := NewDataMessage(fileType, clientId, requestId, data)
	if err := utils.WriteToSocket(conn, msg.Marshal()); err != nil {
		return err
	}

	return nil
}

func ReadMsg(conn net.Conn) (Message, error) {
	buf := make([]byte, HeaderSize)
	if err := utils.ReadFromSocket(conn, &buf, HeaderSize); err != nil {
		return Message{}, fmt.Errorf("failed to read msg from socket: %w", err)
	}

	msgHeader := Header{}
	if err := msgHeader.Unmarshal(buf); err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	payloadBuf := make([]byte, msgHeader.PayloadSize)
	if err := utils.ReadFromSocket(conn, &payloadBuf, int(msgHeader.PayloadSize)); err != nil {
		return Message{}, fmt.Errorf("failed to read payload from socket: %w", err)
	}

	return Message{Header: msgHeader, Payload: payloadBuf}, nil
}

func ReadSyncMsg(conn net.Conn) (Message, error) {
	msg, err := ReadMsg(conn)
	if err != nil {
		return Message{}, err
	}

	if !msg.IsSync() {
		return Message{}, fmt.Errorf("msg isn't of type Sync")
	}

	return msg, nil
}

func ReadAckSyncMsg(conn net.Conn) (Message, error) {
	msg, err := ReadMsg(conn)
	if err != nil {
		return Message{}, err
	}

	if !msg.IsSyncAck() {
		return msg, fmt.Errorf("msg isn't of type SyncAck")
	}

	return msg, nil
}

func ReadStartMsg(conn net.Conn) (Message, error) {
	msg, err := ReadMsg(conn)
	if err != nil {
		return Message{}, err
	}

	if !msg.IsStart() {
		return Message{}, fmt.Errorf("msg isn't of type Start")
	}

	return msg, nil
}
