package message

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/payload"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/utils"
)

type DataType uint8

const (
	Reviews DataType = iota
	Games
)

type DataMessageConfig struct {
	Start    bool
	End      bool
	DataType DataType
	Data     []byte
}

func NewDefaultDataMessage() *Message[*payload.Data] {
	header := &Header{}
	payload := payload.NewData()
	return newMessage(header, payload)
}

func NewDataMessage(dataMsgConf *DataMessageConfig) *Message[*payload.Data] {
	payload := &payload.Data{
		Header: &utils.StreamHeader{
			Type:  uint8(dataMsgConf.DataType),
			Start: utils.GetStartFlag(dataMsgConf.Start),
			End:   utils.GetEndFlag(dataMsgConf.End),
		},
		Payload: &utils.StreamPayload{
			Data: dataMsgConf.Data,
		},
	}
	header := &Header{
		Optype:      Data,
		PayloadSize: uint32(payload.Sizeof()),
	}
	return newMessage(header, payload)
}
