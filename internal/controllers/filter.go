package controllers

import (
	"context"
	"fmt"
	filter2 "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/filter"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/routing"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/pemistahl/lingua-go"
	"log/slog"
)

type Filter struct {
	io         client.IOManager
	router     routing.Router
	usesRouter bool

	filterFunc filter2.FuncFilter
	detector   *lingua.LanguageDetector

	done chan struct{}
}

// TODO(fede) - Lista de Juegos o Reviews a enviar debe devolver los filtros por el routing
// TODO(fede) - Conditional IOManager depending of filter
// TODO(fede)- FilterFunc from env variable
func NewFilter(filter filter2.FuncFilterName) (Filter, error) {
	// Obtains the filter function
	filterFunc, ok := filter2.FilterMap[filter]
	if !ok {
		return Filter{}, fmt.Errorf("unknown filter: %s", filter)
	}

	// If it needs a detector, create one
	var detector *lingua.LanguageDetector
	if filter2.NeedsDecoder(filter) {
		languages := []lingua.Language{
			lingua.English,
			lingua.Spanish,
		}
		detectorObj := lingua.NewLanguageDetectorBuilder().FromLanguages(languages...).Build()
		detector = &detectorObj
	}

	var io client.IOManager
	var router routing.Router
	// Checks filter IO config
	filterIOConfig, ok := filter2.FilterInputsOutputs[filter]
	if !ok {
		return Filter{}, fmt.Errorf("unknown filter IO config: %s", filter)
	}

	if filterIOConfig.UseRouter {
		internalRouter, err := routing.NewRouterById(filterIOConfig.Input)
		if err != nil {
			return Filter{}, fmt.Errorf("failed to create filter router: %w", err)
		}
		router = internalRouter
	} else {
		if err := io.Connect(filterIOConfig.Input, filterIOConfig.Output); err != nil {
			return Filter{}, fmt.Errorf("couldn't create filter io: %w", err)
		}
	}

	return Filter{
		io:         io,
		router:     router,
		usesRouter: filterIOConfig.UseRouter,

		filterFunc: filterFunc,
		detector:   detector,

		done: make(chan struct{}),
	}, nil
}

func (f *Filter) Done() <-chan struct{} {
	return f.done
}

func (f *Filter) Run(ctx context.Context) error {
	consumerCh := f.io.Input.GetConsumer()
	defer func() { f.done <- struct{}{} }()

	for {
		select {
		case delivery := <-consumerCh:
			msgBytes := delivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}

			// Detect type
			if msg.ExpectKind(protocol.Data) {
				// Handle filter
				responseMsg, err := f.filterFunc(msg, f.detector)
				if err != nil {
					return fmt.Errorf("couldn't filter message: %w", err)
				}

				if err = f.io.Write(responseMsg.Marshal(), ""); err != nil {
					return fmt.Errorf("couldn't write data: %w", err)
				}
			} else if msg.ExpectKind(protocol.End) {
				var msgType protocol.DataType
				if msg.HasGameData() {
					msgType = protocol.Games
				} else if msg.HasReviewData() {
					msgType = protocol.Reviews
				}

				newMsg := protocol.NewEndMessage(
					msgType,
					protocol.MessageOptions{
						ClientID:  msg.GetClientID(),
						RequestID: msg.GetRequestID(),
						MessageID: msg.GetMessageID(),
					})
				if err := f.io.Write(newMsg.Marshal(), ""); err != nil {
					return fmt.Errorf("couldn't write end message: %w", err)
				}
				slog.Info("Received End message",
					"clientId", msg.GetClientID(),
					"requestId", msg.GetRequestID(),
					"type", msg.GetRequestID(),
				)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (f *Filter) Close() {
	f.io.Close()
}
