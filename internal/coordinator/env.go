package coordinator

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

const OutputTypeEnv = "OUTPUT_TYPE"
const ExpectedNodes = "EXPECTED_NODES"
const TransactionLogFile = "TRANSACTION_LOG_FILE"

func GetTransactionLogFile() (string, error) {
	value, err := utils.GetFromEnv(TransactionLogFile)
	if err != nil {
		return "", nil
	}

	return *value, nil
}

func GetExpectedNodes() (uint32, error) {
	value, err := utils.GetFromEnvUint(ExpectedNodes)
	if err != nil {
		return 0, err
	}

	return uint32(*value), nil
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
