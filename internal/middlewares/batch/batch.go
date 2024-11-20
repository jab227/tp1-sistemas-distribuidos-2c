package batch

import (
	"bytes"
	"encoding/binary"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/text/message"
)

type Batcher struct {
	messages     []protocol.Message
	lastDelivery amqp091.Delivery
	pos          int
	size         int
}

func NewBatcher(size int) Batcher {
	return Batcher{
		messages: make([]protocol.Message, size),
		size:     size,
	}
}

func (b *Batcher) Push(m protocol.Message, delivery amqp091.Delivery) {
	utils.Assert(b.pos < b.size, "b.pos should be less than b.size")
	b.messages[b.pos] = m
	b.lastDelivery = delivery
	b.pos++
}

func (b Batcher) IsEmpty() bool {
	return b.pos == 0
}

func (b Batcher) IsFull() bool {
	return b.pos == b.size
}

func (b *Batcher) Acknowledge() {
	b.lastDelivery.Ack(true)
	b.pos = 0
}

func (b *Batcher) Batch() []protocol.Message {
	return b.messages[:b.pos]
}

func MarshalBatch(b []protocol.Message) []byte {
	var buf bytes.Buffer
	var bufCount [4]byte
	binary.LittleEndian.PutUint32(bufCount[:], uint32(len(b)))
	buf.Write(bufCount[:])
	for _, m := range b {
		p := m.Marshal()
		binary.LittleEndian.PutUint32(bufCount[:], uint32(len(p)))
		buf.Write(bufCount[:])
		buf.Write(p)
	}
	return buf.Bytes()
}

func UnmarshalBatch(p []byte) ([]protocol.Message, error) {
	count := int(binary.LittleEndian.Uint32(p[0:4]))
	p = p[4:]
	messages := make([]protocol.Message, count)
	for i := 0; i < count; i++ {
		messageSize := int(binary.LittleEndian.Uint32(p[:4]))
		if err := messages[i].Unmarshal(p[4:messageSize]); err != nil {
			return nil, err
		}
		p = p[messageSize+4:]
	}
	return messages, nil
}
