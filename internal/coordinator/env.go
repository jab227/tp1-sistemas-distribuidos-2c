package coordinator

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

const ExpectedGamesEnv = "EXPECTED_GAMES"
const ExpectedReviewsEnv = "EXPECTED_REVIEWS"
const OutputTypeEnv = "OUTPUT_TYPE"

func GetExpectedGames() (int, error) {
	value, err := utils.GetFromEnvInt(ExpectedGamesEnv)
	if err != nil {
		return -1, err
	}

	return int(*value), nil
}

func GetExpectedRevisions() (int, error) {
	value, err := utils.GetFromEnvInt(ExpectedReviewsEnv)
	if err != nil {
		return -1, err
	}

	return int(*value), nil
}

func GetOutputType() (client.OutputType, error) {
	value, err := utils.GetFromEnv(OutputTypeEnv)
	if err != nil {
		return client.NoneOutput, err
	}

	var outputType client.OutputType
	switch *value {
	case "worker":
		outputType = client.OutputWorker
	case "fanout":
		outputType = client.FanoutPublisher
	case "direct":
		outputType = client.Router
	default:
		return client.NoneOutput, fmt.Errorf("invalid output type: %s", *value)
	}

	return outputType, nil
}
