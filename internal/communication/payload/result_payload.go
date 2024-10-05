package payload

import (
	"bytes"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/utils"
)

type Result struct {
	Header  *utils.StreamHeader
	Payload *utils.StreamPayload
}

func NewResult() *Result {
	header := &utils.StreamHeader{}
	payload := &utils.StreamPayload{}
	return &Result{header, payload}
}

func (s *Result) Sizeof() int {
	return s.Header.Sizeof() + s.Payload.Sizeof()
}

func (s *Result) Marshall() []byte {
	header := s.Header.Marshall()
	payload := s.Payload.Marshall()
	return append(header, payload...)
}

func (s *Result) Unmarshall(data []byte) {
	sizeOfHeader := s.Header.Sizeof()
	sizeOfPayload := len(data) - sizeOfHeader

	buff := bytes.NewBuffer(data)
	s.Header.Unmarshall(buff.Next(sizeOfHeader))
	s.Payload.Unmarshall(buff.Next(sizeOfPayload))
}
