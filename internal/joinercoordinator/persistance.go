package joinercoordinator

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

type EndJoinerState struct {
	Games        map[uint32]map[uint32]struct{} // Map with nodes ID, each with a set of clientIds
	Reviews      map[uint32]map[uint32]struct{} // Map with nodes ID, each with a set of clientIds
	FinalReviews map[uint32]map[uint32]struct{} // Map wit nodes ID, each with a set of clientsId

	SentGames        map[uint32]struct{}
	SentReviews      map[uint32]struct{}
	SentFinalReviews map[uint32]struct{}
}

func NewEndJoinerState(numberOfNodes uint32) *EndJoinerState {
	games := make(map[uint32]map[uint32]struct{}, numberOfNodes)
	reviews := make(map[uint32]map[uint32]struct{}, numberOfNodes)
	finalReviews := make(map[uint32]map[uint32]struct{}, numberOfNodes)

	for i := uint32(1); i <= numberOfNodes; i++ {
		games[i] = make(map[uint32]struct{})
		reviews[i] = make(map[uint32]struct{})
		finalReviews[i] = make(map[uint32]struct{})
	}

	return &EndJoinerState{
		Games:            games,
		Reviews:          reviews,
		FinalReviews:     finalReviews,
		SentGames:        make(map[uint32]struct{}),
		SentReviews:      make(map[uint32]struct{}),
		SentFinalReviews: make(map[uint32]struct{}),
	}
}

// TODO - Delete
func (s *EndJoinerState) PrintState() {
	slog.Info("State", "games", s.Games, "reviews", s.Reviews, "finalReviews", s.FinalReviews, "sentGames", s.SentGames, "sentReviews", s.SentReviews, "sentFinalReviews", s.SentFinalReviews)
}

func (s *EndJoinerState) WasSentGame(clientId uint32) bool {
	_, ok := s.SentGames[clientId]
	return ok
}

func (s *EndJoinerState) WasSentReview(clientId uint32) bool {
	_, ok := s.SentReviews[clientId]
	return ok
}

func (s *EndJoinerState) WasSentFinalReview(clientId uint32) bool {
	_, ok := s.SentFinalReviews[clientId]
	return ok
}

func (s *EndJoinerState) SetGameSent(clientId uint32) {
	if _, ok := s.SentGames[clientId]; !ok {
		s.SentGames[clientId] = struct{}{}
	}
	s.SentGames[clientId] = struct{}{}
}

func (s *EndJoinerState) SetReviewSent(clientId uint32) {
	if _, ok := s.SentReviews[clientId]; !ok {
		s.SentReviews[clientId] = struct{}{}
	}
	s.SentReviews[clientId] = struct{}{}
}

func (s *EndJoinerState) SetFinalReview(clientId uint32) {
	if _, ok := s.FinalReviews[clientId]; !ok {
		s.SentFinalReviews[clientId] = struct{}{}
	}
	s.SentFinalReviews[clientId] = struct{}{}
}

func (s *EndJoinerState) AddGame(nodeId uint32, clientId uint32) {
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

func (s *EndJoinerState) AddReview(nodeId uint32, clientId uint32) {
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

func (s *EndJoinerState) AddFinalReview(nodeId uint32, clientId uint32) {
	nodeState, ok := s.FinalReviews[nodeId]
	if !ok {
		nodeState = make(map[uint32]struct{})
	}

	v, ok := nodeState[clientId]
	if !ok {
		v = struct{}{}
		nodeState[clientId] = v
	}

	s.FinalReviews[nodeId] = nodeState
}

func (s *EndJoinerState) AllGamesForId(clientId uint32) bool {
	for _, value := range s.Games {
		if _, ok := value[clientId]; !ok {
			return false
		}
	}

	return true
}

func (s *EndJoinerState) AllReviewsForId(clientId uint32) bool {
	for _, value := range s.Reviews {
		if _, ok := value[clientId]; !ok {
			return false
		}
	}

	return true
}

func (s *EndJoinerState) AllFinalReviewsForId(clientId uint32) bool {
	for _, value := range s.FinalReviews {
		if _, ok := value[clientId]; !ok {
			return false
		}
	}

	return true
}

func (s *EndJoinerState) ResetGame(clientId uint32) {
	for _, value := range s.Games {
		delete(value, clientId)
	}
}

func (s *EndJoinerState) ResetReview(clientId uint32) {
	for _, value := range s.Reviews {
		delete(value, clientId)
	}
}

func (s *EndJoinerState) ResetFinalReview(clientId uint32) {
	for _, value := range s.FinalReviews {
		delete(value, clientId)
	}
}

// persistence functions

func MarshalState(s *EndJoinerState) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(s); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func UnmarshalState(p []byte) (EndJoinerState, error) {
	buf := bytes.NewBuffer(p)
	decoder := gob.NewDecoder(buf)
	var s EndJoinerState
	if err := decoder.Decode(&s); err != nil {
		return EndJoinerState{}, err
	}

	return s, nil
}

func reloadEndJoinerState(fileName string, numberOfNodes uint32) (EndJoinerState, *persistence.TransactionLog, error) {
	EndJoinerState := *NewEndJoinerState(numberOfNodes)

	// Creates the transaction log
	logFile := fmt.Sprintf("../logs/%s", fileName)
	log := persistence.NewTransactionLog(logFile)

	// Read the contents for the transaction log
	logBytes, err := os.ReadFile(logFile)
	if err != nil {
		// If it doesn't exists, returns empty
		if errors.Is(err, os.ErrNotExist) {
			return EndJoinerState, log, nil
		}

		// Otherwise, it is an error
		return EndJoinerState, log, err
	}

	// Unmarshal the transaction log in an specific format
	if err := log.Unmarshal(logBytes); err != nil {
		return EndJoinerState, log, fmt.Errorf("couldn't unmarshal log: %w", err)
	}

	for _, entry := range log.GetLog() {
		switch controllers.TXN(entry.Kind) {
		case controllers.TXNSet:
			EndJoinerState, err = UnmarshalState(entry.Data)
			if err != nil {
				return EndJoinerState, log, err
			}
		}
	}

	return EndJoinerState, log, nil
}
