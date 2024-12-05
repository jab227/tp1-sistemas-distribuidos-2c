package boundary

import (
	"sync"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/cprotocol"
)

type ClientState struct {
	RequestIdCounter uint64

	senderCh chan cprotocol.Message
}

type State struct {
	mu               sync.Mutex
	MessageIdCounter uint32
	clientIdCounter  uint64
	clientsState     map[uint64]ClientState
}

func NewState() *State {
	return &State{
		clientsState:    make(map[uint64]ClientState),
		clientIdCounter: 0,
	}
}

func (s *State) GetNewClientId() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clientIdCounter++
	s.setNewClientId(s.clientIdCounter)
	return s.clientIdCounter
}

func (s *State) setNewClientId(clientId uint64) {
	s.clientsState[clientId] = ClientState{
		RequestIdCounter: 0,
		senderCh:         make(chan cprotocol.Message, 1),
	}
}

func (s *State) GetClientNewRequestId(clientId uint64) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientState := s.clientsState[clientId]
	clientState.RequestIdCounter += 1
	s.clientsState[clientId] = clientState
	return clientState.RequestIdCounter
}

func (s *State) GetClientNewMessageId(clientId uint64) uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MessageIdCounter += 1
	return s.MessageIdCounter
}

func (s *State) GetClientCh(clientId uint64) chan cprotocol.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientState := s.clientsState[clientId]
	return clientState.senderCh
}
