package middleware

import "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"

const RabbitMQHostname = "RABBITMQ_HOSTNAME"
const RabbitMqPort = "RABBITMQ_PORT"
const RabbitMqUsername = "RABBITMQ_USERNAME"
const RabbitMqPassword = "RABBITMQ_PASSWORD"

func GetOptionsFromEnv() (Options, error) {
	hostname, err := utils.GetFromEnv(RabbitMQHostname)
	if err != nil {
		return Options{}, err
	}

	port, err := utils.GetFromEnv(RabbitMqPort)
	if err != nil {
		return Options{}, err
	}

	username, err := utils.GetFromEnv(RabbitMqUsername)
	if err != nil {
		return Options{}, err
	}

	password, err := utils.GetFromEnv(RabbitMqPassword)
	if err != nil {
		return Options{}, err
	}
	return Options{
		Username: *username,
		Password: *password,
		Hostname: *hostname,
		Port:     *port,
	}, err
}
