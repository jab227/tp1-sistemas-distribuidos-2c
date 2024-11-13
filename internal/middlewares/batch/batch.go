package batch

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"github.com/rabbitmq/amqp091-go"
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
	return b.messages
}
