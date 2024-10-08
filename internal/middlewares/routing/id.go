package routing

import "hash"
import "hash/fnv"

type IDRouter struct {
	hasher  hash.Hash64
	idCount int
}

func NewIDRouter(idCount int) IDRouter {
	return IDRouter{
		hasher:  fnv.New64a(),
		idCount: idCount,
	}
}

func (r IDRouter) Select(key string) int {
	r.hasher.Write([]byte(key))
	return int(r.hasher.Sum64()) % r.idCount
}
