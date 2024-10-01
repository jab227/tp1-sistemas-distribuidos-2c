package utils

import (
	"bytes"
	"encoding/binary"
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
	Optype         uint8
	SequenceNumber uint32
	Start          StartFlag
	End            EndFlag
}

func (h *StreamHeader) Sizeof() int {
	optypeSize := int(unsafe.Sizeof(h.Optype))
	sequenceNumberSize := int(unsafe.Sizeof(h.SequenceNumber))
	startSize := int(unsafe.Sizeof(h.Start))
	endSize := int(unsafe.Sizeof(h.End))
	return optypeSize + sequenceNumberSize + startSize + endSize
}

func (h *StreamHeader) Marshall() []byte {
	sizeOfHeader := h.Sizeof()
	buff := make([]byte, 0, sizeOfHeader)
	buff = append(buff, uint8(h.Optype))
	buff = binary.LittleEndian.AppendUint32(buff, h.SequenceNumber)
	buff = append(buff, uint8(h.Start))
	buff = append(buff, uint8(h.End))
	return buff
}

func (h *StreamHeader) Unmarshall(data []byte) {
	optypeSize := int(unsafe.Sizeof(h.Optype))
	sequenceNumberSize := int(unsafe.Sizeof(h.SequenceNumber))
	startSize := int(unsafe.Sizeof(h.Start))
	endSize := int(unsafe.Sizeof(h.End))

	buff := bytes.NewBuffer(data)
	h.Optype = uint8(buff.Next(optypeSize)[0])
	h.SequenceNumber = binary.LittleEndian.Uint32(buff.Next(sequenceNumberSize))
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
