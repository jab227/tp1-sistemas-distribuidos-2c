package controllers

import (
	"bytes"
	"encoding/binary"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"golang.org/x/exp/maps"
)

type TXN uint32

const (
	TXNBatch TXN = iota
	TXNSet
)

type MessageIDSet map[uint32]struct{}

func NewMessageIDSet() MessageIDSet {
	s := make(map[uint32]struct{})
	return s
}

func (m MessageIDSet) Insert(msgs []protocol.Message) {
	for _, msg := range msgs {
		m[msg.GetMessageID()] = struct{}{}
	}
}

func (m MessageIDSet) Contains(messageID uint32) bool {
	_, ok := m[messageID]
	return ok
}

func (m MessageIDSet) Clear() {
	maps.Clear(m)
}

func (m MessageIDSet) Marshal() []byte {
	var buf bytes.Buffer
	var buf4 [4]byte
	binary.LittleEndian.PutUint32(buf4[:], uint32(len(m)))
	buf.Write(buf4[:])
	for k := range m {
		binary.LittleEndian.PutUint32(buf4[:], k)
		buf.Write(buf4[:])
	}
	return buf.Bytes()
}

func (m MessageIDSet) Unmarshal(p []byte) error {
	count := int(binary.LittleEndian.Uint32(p[0:4]))
	p = p[4:]
	for i := 0; i < count; i++ {
		id := binary.LittleEndian.Uint32(p[0:4])
		m[id] = struct{}{}
		p = p[4:]
	}
	return nil
}
