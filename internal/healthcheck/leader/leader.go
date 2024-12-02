package leader

import (
	"net"
	"sync"
)

type State uint8

const (
	StateLeading State = iota
	StateElecting
	StateMonitoring
)

type ElectionServiceConfig struct {
	NeighboursIDs       []int
	NeighboursEndpoints []string
	ID                  int
}

type ElectionService struct {
}
