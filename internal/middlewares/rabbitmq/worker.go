package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type WorkerQueueConfig struct {
	Name          string
	Timeout       uint
	PrefetchCount int
}

type WorkerQueue struct {
	ch       *amqp091.Channel
	q        *amqp091.Queue
	Consumer <-chan amqp091.Delivery
	Config   WorkerQueueConfig
}

func (wq *WorkerQueue) Connect(conn *amqp091.Connection, config WorkerQueueConfig) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	q, err := ch.QueueDeclare(config.Name, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.Qos(config.PrefetchCount, 0, false)
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
	wq.Config = config
	return nil
}

func (wq *WorkerQueue) Write(p []byte) (n int, err error) {
	ct, cancel := context.WithTimeout(
		context.Background(),
		time.Second*time.Duration(wq.Config.Timeout),
	)
	defer cancel()

	err = wq.ch.PublishWithContext(
		ct,
		"",
		wq.Config.Name,
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        p,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to publish message: %w", err)
	}

	return len(p), nil
}

func (wq *WorkerQueue) Read() amqp091.Delivery {
	msg := <-wq.Consumer
	return msg
}

func (wq *WorkerQueue) Close() error {
	return wq.ch.Close()
}
