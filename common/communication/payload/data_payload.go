package payload

import (
	"bytes"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/utils"
)

type Data struct {
	Header  *utils.StreamHeader
	Payload *utils.StreamPayload
}

func NewData() *Data {
	header := &utils.StreamHeader{}
	payload := &utils.StreamPayload{}
	return &Data{header, payload}
}

func (s *Data) Sizeof() int {
	return s.Header.Sizeof() + s.Payload.Sizeof()
}

func (s *Data) Marshall() []byte {
	header := s.Header.Marshall()
	payload := s.Payload.Marshall()
	return append(header, payload...)
}

func (s *Data) Unmarshall(data []byte) {
	sizeOfHeader := s.Header.Sizeof()
	sizeOfPayload := len(data) - sizeOfHeader

	buff := bytes.NewBuffer(data)
	s.Header.Unmarshall(buff.Next(sizeOfHeader))
	s.Payload.Unmarshall(buff.Next(sizeOfPayload))
}
