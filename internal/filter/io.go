package filter

import "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"

var FilterInputsOutputs map[string]struct {
	Input     client.InputType
	Output    client.OutputType
	UseRouter bool
} = map[string]struct {
	Input     client.InputType
	Output    client.OutputType
	UseRouter bool
}{
	IndieFilter:    {Input: client.DirectSubscriber, Output: client.DirectPublisher, UseRouter: true},
	ActionFilter:   {Input: client.DirectSubscriber, Output: client.DirectPublisher, UseRouter: true},
	DecadeFilter:   {Input: client.DirectSubscriber, Output: client.OutputWorker, UseRouter: false},
	PositiveFilter: {Input: client.DirectSubscriber, Output: client.DirectPublisher, UseRouter: true},
	NegativeFilter: {Input: client.DirectSubscriber, Output: client.DirectPublisher, UseRouter: true},
	EnglishFilter:  {Input: client.DirectSubscriber, Output: client.DirectPublisher, UseRouter: true},
}
