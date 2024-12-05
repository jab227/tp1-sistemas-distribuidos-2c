package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type WorkerQueueConfig struct {
	Name          string
	Timeout       uint32
	PrefetchCount int
}

type WorkerQueue struct {
	ch     *amqp091.Channel
	q      *amqp091.Queue
	Config WorkerQueueConfig
}

func NewWorkerQueue(config WorkerQueueConfig) *WorkerQueue {
	return &WorkerQueue{Config: config}
}

func (wq *WorkerQueue) Connect(conn *Connection) error {
	ch, err := conn.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	q, err := ch.QueueDeclare(wq.Config.Name, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.Qos(wq.Config.PrefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	wq.ch = ch
	wq.q = &q
	return nil
}

func (wq *WorkerQueue) Write(p []byte, tag string) error {
	ct, cancel := context.WithTimeout(
		context.Background(),
		time.Second*time.Duration(wq.Config.Timeout),
	)
	defer cancel()

	err := wq.ch.PublishWithContext(
		ct,
		"",
		wq.Config.Name,
		false,
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "text/plain",
			Body:         p,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// TODO(fede) - Handle panic
func (wq *WorkerQueue) GetConsumer() <-chan amqp091.Delivery {
	consumer, err := wq.ch.Consume(
		wq.q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic("failed to create consumer")
	}

	return consumer
}

func (wq *WorkerQueue) Close() error {
	return wq.ch.Close()
}
