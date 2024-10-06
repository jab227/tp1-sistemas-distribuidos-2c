package rabbitmq

import "github.com/rabbitmq/amqp091-go"

type InputHandler interface {
	Connect(conn *Connection) error
	Read() amqp091.Delivery
	GetConsumer() <-chan amqp091.Delivery
	Close() error
}

type OutputHandler interface {
	Connect(conn *Connection) error
	Write(msg []byte, tag string) error
	Close() error
}
