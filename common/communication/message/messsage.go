package message

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/utils"
)

type Optype uint8

const (
	Data Optype = iota
	Result
	Sync
	SyncAck
)

type Header struct {
	Optype      Optype
	ClientId    uint32
	RequestId   uint32
	PayloadSize uint32
}

type Message[T utils.Marshallable] struct {
	Header  *Header
	Payload T
}

func newMessage[T utils.Marshallable](Header *Header, payload T) *Message[T] {
	return &Message[T]{Header, payload}
}

func (m *Message[T]) Marshall() []byte {
	Header := m.marshallHeader()
	payload := m.marshallPayload()
	return append(Header, payload...)
}

func (m *Message[T]) marshallHeader() []byte {
	sizeOfHeader := m.SizeofHeader()
	buff := make([]byte, 0, sizeOfHeader)
	buff = append(buff, uint8(m.Header.Optype))
	buff = binary.LittleEndian.AppendUint32(buff, m.Header.ClientId)
	buff = binary.LittleEndian.AppendUint32(buff, m.Header.RequestId)
	buff = binary.LittleEndian.AppendUint32(buff, m.Header.PayloadSize)
	return buff
}

func (m *Message[T]) marshallPayload() []byte {
	return m.Payload.Marshall()
}

func (m *Message[T]) Unmarshall(data []byte) {
	sizeOfHeader := m.SizeofHeader()
	sizeOfPayload := len(data) - sizeOfHeader

	buff := bytes.NewBuffer(data)
	m.UnmarshallHeader(buff.Next(sizeOfHeader))
	m.UnmarshallPayload(buff.Next(sizeOfPayload))
}

func (m *Message[T]) UnmarshallHeader(data []byte) {
	optypeSize := int(unsafe.Sizeof(m.Header.Optype))
	clientIdSize := int(unsafe.Sizeof(m.Header.ClientId))
	requestIdSize := int(unsafe.Sizeof(m.Header.RequestId))
	payloadSizeSize := int(unsafe.Sizeof(m.Header.PayloadSize))

	buff := bytes.NewBuffer(data)
	m.Header.Optype = Optype(buff.Next(optypeSize)[0])
	m.Header.ClientId = binary.LittleEndian.Uint32(buff.Next(clientIdSize))
	m.Header.RequestId = binary.LittleEndian.Uint32(buff.Next(requestIdSize))
	m.Header.PayloadSize = binary.LittleEndian.Uint32(buff.Next(payloadSizeSize))
}

func (m *Message[T]) UnmarshallPayload(data []byte) {
	m.Payload.Unmarshall(data)
}

func (m *Message[T]) SizeofHeader() int {
	optypeSize := int(unsafe.Sizeof(m.Header.Optype))
	clientIdSize := int(unsafe.Sizeof(m.Header.ClientId))
	requestIdSize := int(unsafe.Sizeof(m.Header.RequestId))
	payloadSizeSize := int(unsafe.Sizeof(m.Header.PayloadSize))
	return optypeSize + clientIdSize + requestIdSize + payloadSizeSize
}
