package message

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/payload"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/utils"
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
	Start      bool
	End        bool
	ResultType ResultType
	Data       []byte
}

func NewDefaultResultMessage() *Message[*payload.Result] {
	header := &Header{}
	payload := payload.NewResult()
	return newMessage(header, payload)
}

func NewResultMessage(resultMsgConf *ResultMessageConfig) *Message[*payload.Result] {
	payload := &payload.Result{
		Header: &utils.StreamHeader{
			Type:  uint8(resultMsgConf.ResultType),
			Start: utils.GetStartFlag(resultMsgConf.Start),
			End:   utils.GetEndFlag(resultMsgConf.End),
		},
		Payload: &utils.StreamPayload{
			Data: resultMsgConf.Data,
		},
	}
	header := &Header{
		Optype:      Result,
		PayloadSize: uint32(payload.Sizeof()),
	}

	return newMessage(header, payload)
}
