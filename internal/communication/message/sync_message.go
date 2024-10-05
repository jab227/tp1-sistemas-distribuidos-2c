package message

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/payload"
)

func NewDefaultSyncMessage() *Message[*payload.Empty] {
	header := &Header{}
	payload := payload.NewEmpty()
	return newMessage(header, payload)
}

func NewSyncMessage() *Message[*payload.Empty] {
	payload := &payload.Empty{}
	header := &Header{
		Optype:      Sync,
		ClientId:    0,
		RequestId:   0,
		PayloadSize: uint32(payload.Sizeof()),
	}
	return newMessage(header, payload)
}
