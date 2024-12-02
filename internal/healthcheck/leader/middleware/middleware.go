package middleware

import (
	"context"
	"fmt"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

const exchangeName = "leader-exchange"

type Message interface {
	GetBody() []byte
	Acknowledge()
}

type rabbitMessage amqp.Delivery

func (r rabbitMessage) GetBody() []byte {
	return amqp.Delivery(r).Body
}

func (r rabbitMessage) Acknowledge() {
	amqp.Delivery(r).Ack(false)
}

type LeaderMiddleware struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

type Options struct {
	Username, Password, Hostname, Port string
}

func NewLeaderMiddleware(id int, options *Options) (*LeaderMiddleware, error) {
	conn, err := amqp.Dial(
		fmt.Sprintf("amqp://%s:%s@%s:%s/",
			options.Username,
			options.Password,
			options.Hostname,
			options.Port,
		))
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type
		false,        // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	name := fmt.Sprintf("leader-queue-%d", id)
	q, err := ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}
	err = ch.QueueBind(
		q.Name,           // queue name
		strconv.Itoa(id), // routing key
		exchangeName,     // exchange
		false,
		nil)
	if err != nil {
		return nil, err
	}
	return &LeaderMiddleware{
		conn:  conn,
		ch:    ch,
		queue: q,
	}, nil
}

func (l *LeaderMiddleware) Close() {
	l.conn.Close()
}

func (l *LeaderMiddleware) Write(ctx context.Context, p []byte, key int) error {
	keyStr := strconv.Itoa(key)
	return l.ch.PublishWithContext(ctx,
		exchangeName, // exchange
		keyStr,       // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        p,
		})
}

func (l *LeaderMiddleware) Reader() (<-chan Message, error) {
	rabbitMsgs, err := l.ch.Consume(
		l.queue.Name, // queue
		"",           // consumer
		false,        // auto ack
		false,        // exclusive
		false,        // no local
		false,        // no wait
		nil,          // args
	)
	if err != nil {
		return nil, err
	}
	messages := make(chan Message, 1)
	go func() {
		for delivery := range rabbitMsgs {
			messages <- rabbitMessage(delivery)
		}
	}()
	return messages, nil
}
