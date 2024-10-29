package boundary

import "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"

const ServerPortEnv = "SERVER_PORT"
const ServerHostEnv = "SERVER_HOST"

type Config struct {
	ServiceHost string
	ServicePort int
}

func GetBoundaryConfigFromEnv() (*Config, error) {
	port, err := utils.GetFromEnvInt(ServerPortEnv)
	if err != nil {
		return nil, err
	}

	host, err := utils.GetFromEnv(ServerHostEnv)
	if err != nil {
		return nil, err
	}

	return &Config{
		ServiceHost: *host,
		ServicePort: int(*port),
	}, nil
}
