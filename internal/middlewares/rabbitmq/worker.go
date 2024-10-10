package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type WorkerQueueConfig struct {
	Name          string
	Timeout       uint8
	PrefetchCount int
}

type WorkerQueue struct {
	ch       *amqp091.Channel
	q        *amqp091.Queue
	Consumer <-chan amqp091.Delivery
	Config   WorkerQueueConfig
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

	consumer, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	wq.ch = ch
	wq.q = &q
	wq.Consumer = consumer
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

func (wq *WorkerQueue) GetConsumer() <-chan amqp091.Delivery {
	return wq.Consumer
}

func (wq *WorkerQueue) Read() amqp091.Delivery {
	msg := <-wq.Consumer
	return msg
}

func (wq *WorkerQueue) Close() error {
	return wq.ch.Close()
}
