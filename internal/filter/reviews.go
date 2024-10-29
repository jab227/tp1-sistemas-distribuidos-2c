package filter

import (
	"errors"
	"log/slog"

	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/pemistahl/lingua-go"
)

type FuncFilterReviews func(msg protocol.Message, detector *lingua.LanguageDetector) ([]models.Review, error)

const (
	PositiveFilter string = "positiveFilter"
	NegativeFilter string = "negativeFilter"
	EnglishFilter  string = "englishFilter"
)

var FilterReviewsMap map[string]FuncFilterReviews = map[string]FuncFilterReviews{
	PositiveFilter: FilterByPositiveScore,
	NegativeFilter: FilterByNegativeScore,
	EnglishFilter:  FilterByEnglish,
}

func NeedsDecoder(name string) bool {
	return name == EnglishFilter
}

func FilterByPositiveScore(msg protocol.Message, detector *lingua.LanguageDetector) ([]models.Review, error) {
	var listOfPassed []models.Review

	if !msg.HasReviewData() {
		return []models.Review{}, errors.New("expected review data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		review := models.ReadReview(&element)
		if review.Score == models.Positive {
			listOfPassed = append(listOfPassed, review)
		}
	}

	return listOfPassed, nil
}

func FilterByNegativeScore(msg protocol.Message, detector *lingua.LanguageDetector) ([]models.Review, error) {
	var listOfPassed []models.Review

	if !msg.HasReviewData() {
		slog.Debug("Shouldn't happen")
		return []models.Review{}, errors.New("expected review data")
	}
	elements := msg.Elements()
	for _, element := range elements.Iter() {
		review := models.ReadReview(&element)
		if review.Score == models.Negative {
			listOfPassed = append(listOfPassed, review)
		}
	}

	return listOfPassed, nil
}

func FilterByEnglish(msg protocol.Message, detector *lingua.LanguageDetector) ([]models.Review, error) {
	var listOfPassed []models.Review

	if !msg.HasReviewData() {
		return []models.Review{}, errors.New("expected review data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		review := models.ReadReview(&element)
		language, exist := (*detector).DetectLanguageOf(review.Text)
		if exist && language == lingua.English {
			listOfPassed = append(listOfPassed, review)
		}
	}

	return listOfPassed, nil
}
