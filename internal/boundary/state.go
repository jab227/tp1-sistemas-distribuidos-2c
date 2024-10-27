package boundary

import "sync"

type ClientState struct {
	RequestIdCounter uint64
	MessageIdCounter uint32
}

type State struct {
	mu sync.Mutex

	clientIdCounter uint64
	clientsState    map[uint64]ClientState
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
	return s.clientIdCounter
}

func (s *State) SetNewClientId(clientId uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clientsState[clientId] = ClientState{
		RequestIdCounter: 0,
		MessageIdCounter: 0,
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

	clientState := s.clientsState[clientId]
	clientState.MessageIdCounter += 1
	s.clientsState[clientId] = clientState

	return clientState.MessageIdCounter
}
