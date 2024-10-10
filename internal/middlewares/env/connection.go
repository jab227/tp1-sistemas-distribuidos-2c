package env

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

const RabbitMQHostname = "RABBITMQ_HOSTNAME"
const RabbitMqPort = "RABBITMQ_PORT"
const RabbitMqUsername = "RABBITMQ_USERNAME"
const RabbitMqPassword = "RABBITMQ_PASSWORD"

func GetConnection() (*rabbitmq.Connection, error) {
	hostname, err := utils.GetFromEnv(RabbitMQHostname)
	if err != nil {
		return nil, err
	}

	port, err := utils.GetFromEnv(RabbitMqPort)
	if err != nil {
		return nil, err
	}

	username, err := utils.GetFromEnv(RabbitMqUsername)
	if err != nil {
		return nil, err
	}

	password, err := utils.GetFromEnv(RabbitMqPassword)
	if err != nil {
		return nil, err
	}

	connection := rabbitmq.Connection{}
	err = connection.Connect(*hostname, *port, *username, *password)
	if err != nil {
		return nil, err
	}

	return &connection, nil
}
