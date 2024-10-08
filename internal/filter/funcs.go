package filter

import (
	"errors"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/pemistahl/lingua-go"
	"strings"
)

type FuncFilter func(msg protocol.Message, detector *lingua.LanguageDetector) (protocol.Message, error)
type FuncFilterName string

const (
	IndieFilter    FuncFilterName = "indieFilter"
	ActionFilter   FuncFilterName = "actionFilter"
	DecadeFilter   FuncFilterName = "decadeFilter"
	PositiveFilter FuncFilterName = "positiveFilter"
	NegativeFilter FuncFilterName = "negativeFilter"
	EnglishFilter  FuncFilterName = "englishFilter"
)

var FilterMap map[FuncFilterName]FuncFilter = map[FuncFilterName]FuncFilter{
	IndieFilter:    FilterByGenreIndie,
	ActionFilter:   FilterByGenreAction,
	DecadeFilter:   FilterByDecade,
	PositiveFilter: FilterByPositiveScore,
	NegativeFilter: FilterByNegativeScore,
	EnglishFilter:  FilterByEnglish,
}

// TODO - Improve how it is selected router or iomanager
var FilterInputsOutputs map[FuncFilterName]struct {
	Input     client.InputType
	Output    client.OutputType
	UseRouter bool
} = map[FuncFilterName]struct {
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

func NeedsDecoder(name FuncFilterName) bool {
	return name == EnglishFilter
}

func FilterByGenreIndie(msg protocol.Message, detector *lingua.LanguageDetector) (protocol.Message, error) {
	var listOfPassed []models.Game

	if !msg.HasGameData() {
		return protocol.Message{}, fmt.Errorf("expected game data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		if strings.ToLower(game.Genres) == "indie" {
			listOfPassed = append(listOfPassed, game)
		}
	}

	payloadBufferPassed := protocol.NewPayloadBuffer(len(listOfPassed))
	for _, game := range listOfPassed {
		payloadBufferPassed.BeginPayloadElement()
		game.BuildPayload(payloadBufferPassed)
		payloadBufferPassed.EndPayloadElement()
	}

	responseMsgPassed := protocol.NewDataMessage(
		protocol.Games,
		payloadBufferPassed.Bytes(),
		protocol.MessageOptions{
			ClientID:  msg.GetClientID(),
			RequestID: msg.GetRequestID(),
			MessageID: msg.GetMessageID(),
		},
	)

	return responseMsgPassed, nil
}

func FilterByGenreAction(msg protocol.Message, detector *lingua.LanguageDetector) (protocol.Message, error) {
	var listOfPassed []models.Game

	if !msg.HasGameData() {
		return protocol.Message{}, fmt.Errorf("expected game data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		if strings.ToLower(game.Genres) == "action" {
			listOfPassed = append(listOfPassed, game)
		}
	}

	payloadBufferPassed := protocol.NewPayloadBuffer(len(listOfPassed))
	for _, game := range listOfPassed {
		payloadBufferPassed.BeginPayloadElement()
		game.BuildPayload(payloadBufferPassed)
		payloadBufferPassed.EndPayloadElement()
	}

	responseMsgPassed := protocol.NewDataMessage(
		protocol.Games,
		payloadBufferPassed.Bytes(),
		protocol.MessageOptions{
			ClientID:  msg.GetClientID(),
			RequestID: msg.GetRequestID(),
			MessageID: msg.GetMessageID(),
		},
	)

	return responseMsgPassed, nil
}

func FilterByDecade(msg protocol.Message, detector *lingua.LanguageDetector) (protocol.Message, error) {
	var listOfPassed []models.Game

	if !msg.HasGameData() {
		return protocol.Message{}, errors.New("expected game data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		if game.ReleaseYear <= 2020 && game.ReleaseYear >= 2010 {
			listOfPassed = append(listOfPassed, game)
		}
	}

	payloadBufferPassed := protocol.NewPayloadBuffer(len(listOfPassed))
	for _, game := range listOfPassed {
		payloadBufferPassed.BeginPayloadElement()
		game.BuildPayload(payloadBufferPassed)
		payloadBufferPassed.EndPayloadElement()
	}

	responseMsgPassed := protocol.NewDataMessage(
		protocol.Games,
		payloadBufferPassed.Bytes(),
		protocol.MessageOptions{
			ClientID:  msg.GetClientID(),
			RequestID: msg.GetRequestID(),
			MessageID: msg.GetMessageID(),
		},
	)

	return responseMsgPassed, nil
}

func FilterByPositiveScore(msg protocol.Message, detector *lingua.LanguageDetector) (protocol.Message, error) {
	var listOfPassed []models.Review

	if !msg.HasReviewData() {
		return protocol.Message{}, errors.New("expected review data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		review := models.ReadReview(&element)
		if review.Score == models.Positive {
			listOfPassed = append(listOfPassed, review)
		}
	}

	payloadBufferPassed := protocol.NewPayloadBuffer(len(listOfPassed))
	for _, review := range listOfPassed {
		payloadBufferPassed.BeginPayloadElement()
		review.BuildPayload(payloadBufferPassed)
		payloadBufferPassed.EndPayloadElement()
	}

	responseMsgPassed := protocol.NewDataMessage(
		protocol.Reviews,
		payloadBufferPassed.Bytes(),
		protocol.MessageOptions{
			ClientID:  msg.GetClientID(),
			RequestID: msg.GetRequestID(),
			MessageID: msg.GetMessageID(),
		},
	)

	return responseMsgPassed, nil
}

func FilterByNegativeScore(msg protocol.Message, detector *lingua.LanguageDetector) (protocol.Message, error) {
	var listOfPassed []models.Review

	if !msg.HasReviewData() {
		return protocol.Message{}, errors.New("expected review data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		review := models.ReadReview(&element)
		if review.Score == models.Negative {
			listOfPassed = append(listOfPassed, review)
		}
	}

	payloadBufferPassed := protocol.NewPayloadBuffer(len(listOfPassed))
	for _, review := range listOfPassed {
		payloadBufferPassed.BeginPayloadElement()
		review.BuildPayload(payloadBufferPassed)
		payloadBufferPassed.EndPayloadElement()
	}

	responseMsgPassed := protocol.NewDataMessage(
		protocol.Reviews,
		payloadBufferPassed.Bytes(),
		protocol.MessageOptions{
			ClientID:  msg.GetClientID(),
			RequestID: msg.GetRequestID(),
			MessageID: msg.GetMessageID(),
		},
	)

	return responseMsgPassed, nil
}

func FilterByEnglish(msg protocol.Message, detector *lingua.LanguageDetector) (protocol.Message, error) {
	var listOfPassed []models.Review

	if !msg.HasReviewData() {
		return protocol.Message{}, errors.New("expected review data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		review := models.ReadReview(&element)
		language, exist := (*detector).DetectLanguageOf(review.Text)
		if exist && language == lingua.English {
			listOfPassed = append(listOfPassed, review)
		}
	}

	payloadBufferPassed := protocol.NewPayloadBuffer(len(listOfPassed))
	for _, review := range listOfPassed {
		payloadBufferPassed.BeginPayloadElement()
		review.BuildPayload(payloadBufferPassed)
		payloadBufferPassed.EndPayloadElement()
	}

	responseMsgPassed := protocol.NewDataMessage(
		protocol.Reviews,
		payloadBufferPassed.Bytes(),
		protocol.MessageOptions{
			ClientID:  msg.GetClientID(),
			RequestID: msg.GetRequestID(),
			MessageID: msg.GetMessageID(),
		},
	)

	return responseMsgPassed, nil
}
