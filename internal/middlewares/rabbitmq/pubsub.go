package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type DirectPublisherConfig struct {
	Exchange string
	Timeout  uint32
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
	Exchange      []string
	Queue         string
	Keys          []string
	PrefetchCount int
}

type DirectSubscriber struct {
	ch     *amqp091.Channel
	q      *amqp091.Queue
	Config DirectSubscriberConfig
}

func NewDirectSubscriber(config DirectSubscriberConfig) *DirectSubscriber {
	return &DirectSubscriber{Config: config}
}

func (s *DirectSubscriber) Connect(conn *Connection) error {
	ch, err := conn.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// TODO(fede) - Allow this to be configurable - Improve it
	if s.Config.PrefetchCount > 0 {
		err = ch.Qos(
			s.Config.PrefetchCount, // prefetch count
			0,                      // prefetch size
			false,                  // global
		)
		if err != nil {
			return fmt.Errorf("failed to set QoS: %w", err)
		}

	}

	for _, exchange := range s.Config.Exchange {
		err = ch.ExchangeDeclare(
			exchange,
			"direct",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	q, err := ch.QueueDeclare(
		s.Config.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	for _, key := range s.Config.Keys {
		for _, exchange := range s.Config.Exchange {
			err = ch.QueueBind(
				q.Name,
				key,
				exchange,
				false,
				nil,
			)
			if err != nil {
				return fmt.Errorf("failed to bind queue %s to exchange %s: %w", q.Name, exchange, err)
			}
		}
	}

	s.ch = ch
	s.q = &q
	return nil
}

// TODO(fede) - Handle panic
func (s *DirectSubscriber) GetConsumer() <-chan amqp091.Delivery {
	consumer, err := s.ch.Consume(
		s.q.Name,
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

func (s *DirectSubscriber) Close() error {
	return s.ch.Close()
}
