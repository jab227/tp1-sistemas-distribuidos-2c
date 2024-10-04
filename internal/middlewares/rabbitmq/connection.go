package rabbitmq

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	conn *amqp091.Connection
}

func (c *Connection) Connect(hostname string, port string, username string, password string) error {
	conn, err := amqp091.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, hostname, port))
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %s", err)
	}

	c.conn = conn
	return nil
}

func (c *Connection) GetConnection() *amqp091.Connection {
	return c.conn
}

func (c *Connection) Close() error {
	return c.conn.Close()
}
