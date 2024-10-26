package cprotocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type OperationType byte

const (
	Start OperationType = iota
	Data
	Result
	Sync
	SyncAck
	End
)

type FileType byte

const (
	None FileType = iota
	Games
	Reviews
)

// TODO(fede) - Replace with unsafe
const HeaderSize = 26

type Header struct {
	OpType      OperationType
	FileType    FileType
	ClientId    uint64
	RequestId   uint64
	PayloadSize uint64
}

type Message struct {
	Header  Header
	Payload []byte
}

func (h *Header) Marshal() []byte {
	buffer := bytes.Buffer{}

	buffer.WriteByte(byte(h.OpType))
	buffer.WriteByte(byte(h.FileType))

	clientIdBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(clientIdBuf, h.ClientId)
	buffer.Write(clientIdBuf)

	requestIdBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(requestIdBuf, h.RequestId)
	buffer.Write(requestIdBuf)

	payloadSizeBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(payloadSizeBuf, h.PayloadSize)
	buffer.Write(payloadSizeBuf)

	return buffer.Bytes()
}

func (h *Header) Unmarshal(b []byte) error {
	if len(b) < HeaderSize {
		return errors.New("header too short")
	}

	h.OpType = OperationType(b[0])
	h.FileType = FileType(b[1])

	h.ClientId = binary.BigEndian.Uint64(b[2:10])
	h.RequestId = binary.BigEndian.Uint64(b[10:18])
	h.PayloadSize = binary.BigEndian.Uint64(b[18:26])

	return nil
}

// TODO(fede) - Esta estructura tiene hardcodeado ClientId y RequestId, en el futuro deberia
// poder pasar el cientId para manejo de clientes caidos
// TODO(fede) - Se agrega payload 1 para que no haya errores en el parseo, no se usa el valor

func NewSyncMessage() *Message {
	payload := []byte{1}
	header := Header{
		OpType:      Sync,
		FileType:    None,
		ClientId:    0,
		RequestId:   0,
		PayloadSize: uint64(len(payload)),
	}

	return &Message{
		Header:  header,
		Payload: payload,
	}
}

func NewAckSyncMessage(clientId uint64, requestId uint64) *Message {
	payload := []byte{1}
	header := Header{
		OpType:      SyncAck,
		FileType:    None,
		ClientId:    clientId,
		RequestId:   requestId,
		PayloadSize: uint64(len(payload)),
	}

	return &Message{
		Header:  header,
		Payload: payload,
	}
}

func NewStartMessage(fileType FileType, clientId uint64, requestId uint64) *Message {
	payload := []byte{1}
	header := Header{
		OpType:      Start,
		FileType:    fileType,
		ClientId:    clientId,
		RequestId:   requestId,
		PayloadSize: uint64(len(payload)),
	}

	return &Message{
		Header:  header,
		Payload: payload,
	}
}

func NewEndMessage(fileType FileType, clientId uint64, requestId uint64) *Message {
	payload := []byte{1}
	header := Header{
		OpType:      End,
		FileType:    fileType,
		ClientId:    clientId,
		RequestId:   requestId,
		PayloadSize: uint64(len(payload)),
	}

	return &Message{
		Header:  header,
		Payload: payload,
	}
}

func NewDataMessage(fileType FileType, clientId uint64, requestId uint64, payload []byte) *Message {
	header := Header{
		OpType:      Data,
		FileType:    fileType,
		ClientId:    clientId,
		RequestId:   requestId,
		PayloadSize: uint64(len(payload)),
	}

	return &Message{
		Header:  header,
		Payload: payload,
	}
}

func (m *Message) Marshal() []byte {
	buffer := bytes.Buffer{}

	buffer.Write(m.Header.Marshal())
	buffer.Write(m.Payload)

	return buffer.Bytes()
}

// Helper methods

func (m *Message) IsStart() bool {
	return m.Header.OpType == Start
}

func (m *Message) IsData() bool {
	return m.Header.OpType == Data
}

func (m *Message) IsResult() bool {
	return m.Header.OpType == Result
}

func (m *Message) IsSync() bool {
	return m.Header.OpType == Sync
}

func (m *Message) IsSyncAck() bool {
	return m.Header.OpType == SyncAck
}

func (m *Message) IsEnd() bool {
	return m.Header.OpType == End
}

func (m *Message) IsGamesMsg() bool {
	return m.Header.FileType == Games
}

func (m *Message) IsReviewsMsg() bool {
	return m.Header.FileType == Reviews
}
