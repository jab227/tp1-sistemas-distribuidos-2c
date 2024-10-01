package message

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/payload"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/utils"
)

type ResultType uint8

const (
	Query1 ResultType = iota
	Query2
	Query3
	Query4
	Query5
)

type ResultMessageConfig struct {
	ClientId       uint32
	RequestId      uint32
	SequenceNumber uint32
	Start          bool
	End            bool
	ResultType     ResultType
	Data           []byte
}

func NewDefaultResultMessage() *Message[*payload.Result] {
	header := &Header{}
	payload := payload.NewResult()
	return newMessage(header, payload)
}

func NewResultMessage(resultMsgConf *ResultMessageConfig) *Message[*payload.Result] {
	payload := &payload.Result{
		Header: &utils.StreamHeader{
			Optype:         uint8(resultMsgConf.ResultType),
			SequenceNumber: resultMsgConf.SequenceNumber,
			Start:          utils.GetStartFlag(resultMsgConf.Start),
			End:            utils.GetEndFlag(resultMsgConf.End),
		},
		Payload: &utils.StreamPayload{
			Data: resultMsgConf.Data,
		},
	}
	header := &Header{
		Optype:      Result,
		ClientId:    resultMsgConf.ClientId,
		RequestId:   resultMsgConf.RequestId,
		PayloadSize: uint32(payload.Sizeof()),
	}

	return newMessage(header, payload)
}
