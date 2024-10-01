package message

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/payload"
)

type SyncAckMessageConfig struct {
	ClientId  uint32
	RequestId uint32
}

func NewDefaultSyncAckMessage() *Message[*payload.Empty] {
	header := &Header{}
	payload := payload.NewEmpty()
	return newMessage(header, payload)
}

func NewSyncAckMessage(syncAckMsgConf *SyncAckMessageConfig) *Message[*payload.Empty] {
	payload := &payload.Empty{}
	header := &Header{
		Optype:      SyncAck,
		ClientId:    syncAckMsgConf.ClientId,
		RequestId:   syncAckMsgConf.RequestId,
		PayloadSize: uint32(payload.Sizeof()),
	}
	return newMessage(header, payload)
}
