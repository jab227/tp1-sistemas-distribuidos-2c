package filter

import (
	"errors"
	"fmt"
	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"strings"
)

type FunFilterGames func(msg protocol.Message) ([]models.Game, error)

const (
	IndieFilter  string = "indieFilter"
	ActionFilter string = "actionFilter"
	DecadeFilter string = "decadeFilter"
)

var FilterGamesMap map[string]FunFilterGames = map[string]FunFilterGames{
	IndieFilter:  FilterByGenreIndie,
	ActionFilter: FilterByGenreAction,
	DecadeFilter: FilterByDecade,
}

func FilterByGenreIndie(msg protocol.Message) ([]models.Game, error) {
	var listOfPassed []models.Game

	if !msg.HasGameData() {
		return []models.Game{}, fmt.Errorf("expected game data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		genres := strings.Split(game.Genres, ",")
		listOfPassed = AppendIfContainsGenre(game, listOfPassed, "indie")
	}

	return listOfPassed, nil
}

func FilterByGenreAction(msg protocol.Message) ([]models.Game, error) {
	var listOfPassed []models.Game

	if !msg.HasGameData() {
		return []models.Game{}, fmt.Errorf("expected game data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		listOfPassed = AppendIfContainsGenre(game, listOfPassed, "action")
	}
	return listOfPassed, nil
}

func AppendIfContainsGenre(game models.Game, listOfPassed []models.Game, genre string) []models.Game {
	genres := strings.Split(game.Genres, ",")
	for _, genre := range genres {
		if strings.EqualFold(genre, genre) {
			listOfPassed = append(listOfPassed, game)
		}
	}
	return listOfPassed
}

func FilterByDecade(msg protocol.Message) ([]models.Game, error) {
	var listOfPassed []models.Game

	if !msg.HasGameData() {
		return []models.Game{}, errors.New("expected game data")
	}

	elements := msg.Elements()
	for _, element := range elements.Iter() {
		game := models.ReadGame(&element)
		if game.ReleaseYear <= 2020 && game.ReleaseYear >= 2010 {
			listOfPassed = append(listOfPassed, game)
		}
	}

	return listOfPassed, nil
}
