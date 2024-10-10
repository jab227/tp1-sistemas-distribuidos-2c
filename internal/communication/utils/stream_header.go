package utils

import (
	"bytes"
	"unsafe"
)

type StartFlag uint8

const (
	StartNotSet StartFlag = 0
	StartSet    StartFlag = 1
)

type EndFlag uint8

const (
	EndNotSet EndFlag = 0
	EndSet    EndFlag = 1
)

type StreamHeader struct {
	Type  uint8
	Start StartFlag
	End   EndFlag
}

func (h *StreamHeader) Sizeof() int {
	typeSize := int(unsafe.Sizeof(h.Type))
	startSize := int(unsafe.Sizeof(h.Start))
	endSize := int(unsafe.Sizeof(h.End))
	return typeSize + startSize + endSize
}

func (h *StreamHeader) Marshall() []byte {
	sizeOfHeader := h.Sizeof()
	buff := make([]byte, 0, sizeOfHeader)
	buff = append(buff, uint8(h.Type))
	buff = append(buff, uint8(h.Start))
	buff = append(buff, uint8(h.End))
	return buff
}

func (h *StreamHeader) Unmarshall(data []byte) {
	typeSize := int(unsafe.Sizeof(h.Type))
	startSize := int(unsafe.Sizeof(h.Start))
	endSize := int(unsafe.Sizeof(h.End))

	buff := bytes.NewBuffer(data)
	h.Type = uint8(buff.Next(typeSize)[0])
	h.Start = StartFlag(buff.Next(startSize)[0])
	h.End = EndFlag(buff.Next(endSize)[0])
}

func GetStartFlag(start bool) StartFlag {
	if start {
		return StartSet
	}
	return StartNotSet
}

func GetEndFlag(end bool) EndFlag {
	if end {
		return EndSet
	}
	return EndNotSet
}
