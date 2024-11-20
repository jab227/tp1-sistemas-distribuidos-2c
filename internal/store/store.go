package store

import (
	"bytes"
	"encoding/binary"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type Store[S any] map[uint32]S

func NewStore[S any]() Store[S] {
	return make(map[uint32]S)
}

func (c Store[S]) Set(clientID uint32, state S) {
	c[clientID] = state
}

func (c Store[S]) Get(clientID uint32) (S, bool) {
	state, ok := c[clientID]
	return state, ok
}

func (c Store[S]) Delete(clientID uint32) {
	delete(c, clientID)
}

func (c Store[S]) Contains(clientID uint32) bool {
	_, ok := c[clientID]
	return ok
}
