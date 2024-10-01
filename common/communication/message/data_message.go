package message

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/payload"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/utils"
)

type DataType uint8

const (
	Games DataType = iota
	Reviews
)

type DataMessageConfig struct {
	ClientId       uint32
	RequestId      uint32
	SequenceNumber uint32
	Start          bool
	End            bool
	DataType       DataType
	Data           []byte
}

func NewDefaultDataMessage() *Message[*payload.Data] {
	header := &Header{}
	payload := payload.NewData()
	return newMessage(header, payload)
}

func NewDataMessage(dataMsgConf *DataMessageConfig) *Message[*payload.Data] {
	payload := &payload.Data{
		Header: &utils.StreamHeader{
			Optype:         uint8(dataMsgConf.DataType),
			SequenceNumber: dataMsgConf.SequenceNumber,
			Start:          utils.GetStartFlag(dataMsgConf.Start),
			End:            utils.GetEndFlag(dataMsgConf.End),
		},
		Payload: &utils.StreamPayload{
			Data: dataMsgConf.Data,
		},
	}
	header := &Header{
		Optype:      Data,
		ClientId:    dataMsgConf.ClientId,
		RequestId:   dataMsgConf.RequestId,
		PayloadSize: uint32(payload.Sizeof()),
	}
	return newMessage(header, payload)
}
