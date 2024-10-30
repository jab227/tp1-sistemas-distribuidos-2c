package cprotocol

import (
	"errors"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"net"
)

func SendMsg(conn net.Conn, msg Message) error {
	return utils.WriteToSocket(conn, msg.Marshal())
}

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

func SendStartMsg(conn net.Conn, contentType ContentType, clientId uint64, requestId uint64) error {
	msg := NewStartMessage(contentType, clientId, requestId)
	if err := utils.WriteToSocket(conn, msg.Marshal()); err != nil {
		return err
	}

	return nil
}

func SendEndMsg(conn net.Conn, contentType ContentType, clientId uint64, requestId uint64) error {
	msg := NewEndMessage(contentType, clientId, requestId)
	if err := utils.WriteToSocket(conn, msg.Marshal()); err != nil {
		return err
	}

	return nil
}

func SendDataMsg(conn net.Conn, contentType ContentType, clientId uint64, requestId uint64, data []byte) error {
	msg := NewDataMessage(contentType, clientId, requestId, data)
	if err := utils.WriteToSocket(conn, msg.Marshal()); err != nil {
		return err
	}

	return nil
}

func SendResultMsg(conn net.Conn, contentType ContentType, clientId uint64, requestId uint64, data []byte) error {
	msg := NewResultMessage(contentType, clientId, requestId, data)
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

func ReadDataMsg(conn net.Conn) (Message, error) {
	msg, err := ReadMsg(conn)
	if err != nil {
		return Message{}, err
	}

	if !msg.IsData() {
		return Message{}, errors.New("msg isn't of type Data")
	}

	return msg, nil
}

func ReadResultMsg(conn net.Conn) (Message, error) {
	msg, err := ReadMsg(conn)
	if err != nil {
		return Message{}, err
	}

	if !msg.IsResult() {
		return Message{}, errors.New("msg isn't of type Result")
	}

	return msg, nil
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
