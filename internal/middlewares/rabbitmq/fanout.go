package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type FanoutPublisherConfig struct {
	Exchange string
	Timeout  uint8
}

type FanoutPublisher struct {
	ch     *amqp091.Channel
	Config FanoutPublisherConfig
}

func (p *FanoutPublisher) Connect(conn *Connection, config FanoutPublisherConfig) error {
	ch, err := conn.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		config.Exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	p.ch = ch
	p.Config = config
	return nil
}

func (p *FanoutPublisher) Publish(msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.Config.Timeout))
	defer cancel()

	err := p.ch.PublishWithContext(
		ctx,
		p.Config.Exchange,
		"",
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (p *FanoutPublisher) Close() error {
	return p.ch.Close()
}

type FanoutSubscriberConfig struct {
	Exchange string
	Queue    string
}

type FanoutSubscriber struct {
	ch       *amqp091.Channel
	q        *amqp091.Queue
	Consumer <-chan amqp091.Delivery
	Config   FanoutSubscriberConfig
}

func (s *FanoutSubscriber) Connect(conn *Connection, config FanoutSubscriberConfig) error {
	ch, err := conn.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		config.Exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(
		config.Queue,
		true,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		q.Name,
		"",
		config.Exchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
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
		return fmt.Errorf("failed to declare consumer: %w", err)
	}

	s.ch = ch
	s.q = &q
	s.Consumer = consumer
	s.Config = config
	return nil
}

func (s *FanoutSubscriber) Read() amqp091.Delivery {
	msg := <-s.Consumer
	return msg
}

func (s *FanoutSubscriber) Close() error {
	return s.ch.Close()
}
