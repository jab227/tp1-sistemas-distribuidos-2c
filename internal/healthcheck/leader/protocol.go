package leader

import "encoding/binary"

type ElectionMessageType uint8

const (
	ElectionMessageTypeAnnounceElection ElectionMessageType = iota
	ElectionMessageTypeAlive
	ElectionMessageTypeCoordinator
)

type ElectionMessage struct {
	ID   uint32
	Type ElectionMessageType
}

func (e ElectionMessage) Marshal() []byte {
	buf := make([]byte, 5)
	buf[0] = byte(e.Type)
	binary.LittleEndian.PutUint32(buf[1:], e.ID)
	return buf
}

func (e *ElectionMessage) Unmarshal(p []byte) {
	e.Type = ElectionMessageType(p[0])
	e.ID = binary.LittleEndian.Uint32(p[1:])
}
