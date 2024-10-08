package filter

import "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"

var FilterInputsOutputs map[string]struct {
	Input  client.InputType
	Output client.OutputType
} = map[string]struct {
	Input  client.InputType
	Output client.OutputType
}{
	IndieFilter:    {Input: client.DirectSubscriber, Output: client.Router},
	ActionFilter:   {Input: client.DirectSubscriber, Output: client.Router},
	DecadeFilter:   {Input: client.DirectSubscriber, Output: client.OutputWorker},
	PositiveFilter: {Input: client.DirectSubscriber, Output: client.Router},
	NegativeFilter: {Input: client.DirectSubscriber, Output: client.Router},
	EnglishFilter:  {Input: client.DirectSubscriber, Output: client.Router},
}
