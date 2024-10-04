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

func NewFanoutPublisher(config FanoutPublisherConfig) *FanoutPublisher {
	return &FanoutPublisher{Config: config}
}

func (p *FanoutPublisher) Connect(conn *Connection) error {
	ch, err := conn.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		p.Config.Exchange,
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
	return nil
}

func (p *FanoutPublisher) Write(msg []byte, tag string) error {
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

func NewFanoutSubscriber(config FanoutSubscriberConfig) *FanoutSubscriber {
	return &FanoutSubscriber{Config: config}
}

func (s *FanoutSubscriber) Connect(conn *Connection) error {
	ch, err := conn.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		s.Config.Exchange,
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
		s.Config.Queue,
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
		s.Config.Exchange,
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
	return nil
}

func (s *FanoutSubscriber) Read() amqp091.Delivery {
	msg := <-s.Consumer
	return msg
}

func (s *FanoutSubscriber) Close() error {
	return s.ch.Close()
}
