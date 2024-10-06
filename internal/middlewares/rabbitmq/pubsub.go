package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type DirectPublisherConfig struct {
	Exchange string
	Timeout  uint8
}

type DirectPublisher struct {
	ch     *amqp091.Channel
	Config DirectPublisherConfig
}

func NewDirectPublisher(config DirectPublisherConfig) *DirectPublisher {
	return &DirectPublisher{Config: config}
}

func (p *DirectPublisher) Connect(conn *Connection) error {
	ch, err := conn.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		p.Config.Exchange,
		"direct",
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

func (p *DirectPublisher) Write(msg []byte, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.Config.Timeout))
	defer cancel()

	err := p.ch.PublishWithContext(
		ctx,
		p.Config.Exchange,
		key,
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

func (p *DirectPublisher) Close() error {
	return p.ch.Close()
}

type DirectSubscriberConfig struct {
	Exchange string
	Queue    string
	Keys     []string
}

type DirectSubscriber struct {
	ch       *amqp091.Channel
	q        *amqp091.Queue
	Consumer <-chan amqp091.Delivery
	Config   DirectSubscriberConfig
}

func NewDirectSubscriber(config DirectSubscriberConfig) *DirectSubscriber {
	return &DirectSubscriber{Config: config}
}

func (s *DirectSubscriber) Connect(conn *Connection) error {
	ch, err := conn.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		s.Config.Exchange,
		"direct",
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

	for _, key := range s.Config.Keys {
		err = ch.QueueBind(
			q.Name,
			key,
			s.Config.Exchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
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
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	s.ch = ch
	s.q = &q
	s.Consumer = consumer
	return nil
}

func (s *DirectSubscriber) Read() amqp091.Delivery {
	msg := <-s.Consumer
	return msg
}

func (s *DirectSubscriber) GetConsumer() <-chan amqp091.Delivery {
	return s.Consumer
}

func (s *DirectSubscriber) Close() error {
	return s.ch.Close()
}
