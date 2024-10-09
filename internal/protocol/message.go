package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type MessageType byte
type DataType byte
type QueryNumber byte

func (m MessageType) String() string {
	switch m {
	case Data:
		return "data message"
	case End:
		return "end message"
	case Results:
		return "results message"
	default:
		utils.Assertf(false, "unexpected protocol.MessageType: %#v", m)
		return "[unknown message]"
	}
}

const (
	Games DataType = iota
	Reviews
)

const (
	End     MessageType = 0 // 0b00
	Data    MessageType = 1 // 0b01
	Results MessageType = 2 // 0b10
)

const (
	Query1 QueryNumber = (1 << 3)
	Query2 QueryNumber = (1 << 4)
	Query3 QueryNumber = (1 << 5)
	Query4 QueryNumber = (1 << 6)
	Query5 QueryNumber = (1 << 7)
)

type Message struct {
	messageType MessageType
	messageID   uint32
	clientID    uint32
	requestID   uint32
	payloadSize uint32
	payload     []byte
}

type MessageOptions struct {
	MessageID uint32
	ClientID  uint32
	RequestID uint32
}

func NewEndMessage(d DataType, opts MessageOptions) Message {
	messageType := End
	if d == Games {
		messageType |= 0x04
	}

	return Message{
		messageType: messageType,
		messageID:   opts.MessageID,
		clientID:    opts.ClientID,
		requestID:   opts.RequestID,
	}
}

func NewDataMessage(d DataType, payload []byte, opts MessageOptions) Message {
	messageType := Data
	if d == Games {
		messageType |= 0x04
	}
	return Message{
		messageType: messageType,
		messageID:   opts.MessageID,
		clientID:    opts.ClientID,
		requestID:   opts.RequestID,
		payloadSize: uint32(len(payload)),
		payload:     payload,
	}
}

func NewResultsMessage(q QueryNumber, payload []byte, opts MessageOptions) Message {
	messageType := Results
	qb := byte(q)
	messageType |= MessageType(qb)
	return Message{
		messageType: messageType,
		messageID:   opts.MessageID,
		clientID:    opts.ClientID,
		requestID:   opts.RequestID,
		payloadSize: uint32(len(payload)),
		payload:     payload,
	}
}

func (m Message) ExpectKind(kind MessageType) bool {
	b := byte(m.messageType)
	b &= 0x03 // 0b11
	return MessageType(b) == kind

}

func (m Message) GetMessageType() MessageType {
	b := byte(m.messageType)
	b &= 0x03
	return MessageType(b)
}

func (m Message) GetMessageID() uint32 {
	return m.messageID
}

func (m Message) GetClientID() uint32 {
	return m.clientID
}

func (m Message) GetRequestID() uint32 {
	return m.requestID
}

func (m Message) HasGameData() bool {
	utils.Assert(m.ExpectKind(Data) || m.ExpectKind(End), "the payload must be data")
	return m.messageType>>2 == 1
}

func (m Message) HasReviewData() bool {
	utils.Assert(m.ExpectKind(Data) || m.ExpectKind(End), "the payload must be  data")
	return m.messageType>>2 == 0
}

func (m Message) GetQueryNumber() int {
	utils.Assert(m.ExpectKind(Results), "the payload must be  data")
	b := byte(m.messageType)
	if (b & byte(Query1)) == byte(Query1) {
		return 1
	}
	if (b & byte(Query2)) == byte(Query2) {
		return 2
	}
	if (b & byte(Query3)) == byte(Query3) {
		return 3
	}
	if (b & byte(Query4)) == byte(Query4) {
		return 4
	}
	if (b & byte(Query5)) == byte(Query5) {
		return 5
	}
	utils.Assert(false, "malformed header type")
	return -1
}

func (m Message) Marshal() []byte {
	var buf bytes.Buffer
	buf4 := make([]byte, 4)

	buf.WriteByte(byte(m.messageType))

	binary.LittleEndian.PutUint32(buf4, m.messageID)
	buf.Write(buf4)
	binary.LittleEndian.PutUint32(buf4, m.clientID)
	buf.Write(buf4)
	binary.LittleEndian.PutUint32(buf4, m.requestID)
	buf.Write(buf4)

	binary.LittleEndian.PutUint32(buf4, m.payloadSize)
	buf.Write(buf4)

	utils.Assert(m.payloadSize == uint32(len(m.payload)), "the sizes must be equal")

	buf.Write(m.payload)
	return buf.Bytes()
}

func (m *Message) Unmarshal(p []byte) error {
	if len(p) == 0 {
		return fmt.Errorf("invalid message: empty")
	}

	maskedMessageType := MessageType(p[0] & 0x3)
	if maskedMessageType != Data && maskedMessageType != Results && maskedMessageType != End {
		return fmt.Errorf("invalid message: unknown message type")
	}
	m.messageType = MessageType(p[0])
	m.messageID = binary.LittleEndian.Uint32(p[1:5])
	m.clientID = binary.LittleEndian.Uint32(p[5:9])
	m.requestID = binary.LittleEndian.Uint32(p[9:13])
	m.payloadSize = binary.LittleEndian.Uint32(p[13:17])
	m.payload = p[17:]
	utils.Assertf(len(m.payload) == int(m.payloadSize), "the payload size should be equal to the len of the payload: %d != %d", len(m.payload), int(m.payloadSize))
	return nil
}

func (m *Message) Elements() *PayloadElements {
	elements, _ := newPayloadElements(m.payload)
	return elements
}

func (m *Message) SetQueryResult(q QueryNumber) {
	m.messageType |= MessageType(q)
}
