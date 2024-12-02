package coordinator

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/controllers"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/persistence"
	"log/slog"
	"os"
)

type EndState struct {
	Games   map[uint32]map[uint32]struct{} // Map with nodes ID, each with a set of clientIds
	Reviews map[uint32]map[uint32]struct{} // Map with nodes ID, each with a set of clientIds

	SentGames   map[uint32]struct{}
	SentReviews map[uint32]struct{}
}

func NewEndState(numberOfNodes uint32) *EndState {
	games := make(map[uint32]map[uint32]struct{}, numberOfNodes)
	reviews := make(map[uint32]map[uint32]struct{}, numberOfNodes)

	for i := uint32(1); i <= numberOfNodes; i++ {
		games[i] = make(map[uint32]struct{})
		reviews[i] = make(map[uint32]struct{})
	}

	return &EndState{
		Games:       games,
		Reviews:     reviews,
		SentGames:   make(map[uint32]struct{}),
		SentReviews: make(map[uint32]struct{}),
	}
}

// TODO - Delete
func (s *EndState) PrintState() {
	slog.Info("State", "games", s.Games, "reviews", s.Reviews, "sentGames", s.SentGames, "sentReviews", s.SentReviews)
}

func (s *EndState) WasSentGame(clientId uint32) bool {
	_, ok := s.SentGames[clientId]
	return ok
}

func (s *EndState) WasSentReview(clientId uint32) bool {
	_, ok := s.SentReviews[clientId]
	return ok
}

func (s *EndState) SetGameSent(clientId uint32) {
	if _, ok := s.SentGames[clientId]; !ok {
		s.SentGames[clientId] = struct{}{}
	}
	s.SentGames[clientId] = struct{}{}
}

func (s *EndState) SetReviewSent(clientId uint32) {
	if _, ok := s.SentReviews[clientId]; !ok {
		s.SentReviews[clientId] = struct{}{}
	}
	s.SentReviews[clientId] = struct{}{}
}

func (s *EndState) AddGame(nodeId uint32, clientId uint32) {
	nodeState, ok := s.Games[nodeId]
	if !ok {
		nodeState = make(map[uint32]struct{})
	}

	v, ok := nodeState[clientId]
	if !ok {
		v = struct{}{}
		nodeState[clientId] = v
	}

	s.Games[nodeId] = nodeState
}

func (s *EndState) AddReview(nodeId uint32, clientId uint32) {
	nodeState, ok := s.Reviews[nodeId]
	if !ok {
		nodeState = make(map[uint32]struct{})
	}

	v, ok := nodeState[clientId]
	if !ok {
		v = struct{}{}
		nodeState[clientId] = v
	}

	s.Reviews[nodeId] = nodeState
}

func (s *EndState) AllGamesForId(clientId uint32) bool {
	for _, value := range s.Games {
		if _, ok := value[clientId]; !ok {
			return false
		}
	}

	return true
}

func (s *EndState) AllReviewsForId(clientId uint32) bool {
	for _, value := range s.Reviews {
		if _, ok := value[clientId]; !ok {
			return false
		}
	}

	return true
}

func (s *EndState) ResetGame(clientId uint32) {
	for _, value := range s.Games {
		delete(value, clientId)
	}
}

func (s *EndState) ResetReview(clientId uint32) {
	for _, value := range s.Reviews {
		delete(value, clientId)
	}
}

// persistence functions

func MarshalState(s *EndState) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(s); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func UnmarshalState(p []byte) (EndState, error) {
	buf := bytes.NewBuffer(p)
	decoder := gob.NewDecoder(buf)
	var s EndState
	if err := decoder.Decode(&s); err != nil {
		return EndState{}, err
	}

	return s, nil
}

func reloadEndState(fileName string, numberOfNodes uint32) (EndState, *persistence.TransactionLog, error) {
	endState := *NewEndState(numberOfNodes)

	// Creates the transaction log
	logFile := fmt.Sprintf("../logs/%s", fileName)
	log := persistence.NewTransactionLog(logFile)

	// Read the contents for the transaction log
	logBytes, err := os.ReadFile(logFile)
	if err != nil {
		// If it doesn't exists, returns empty
		if errors.Is(err, os.ErrNotExist) {
			return endState, log, nil
		}

		// Otherwise, it is an error
		return endState, log, err
	}

	// Unmarshal the transaction log in an specific format
	if err := log.Unmarshal(logBytes); err != nil {
		return endState, log, fmt.Errorf("couldn't unmarshal log: %w", err)
	}

	for _, entry := range log.GetLog() {
		switch controllers.TXN(entry.Kind) {
		case controllers.TXNSet:
			endState, err = UnmarshalState(entry.Data)
			if err != nil {
				return endState, log, err
			}
		}
	}

	return endState, log, nil
}
