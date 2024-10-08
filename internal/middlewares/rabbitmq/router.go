package rabbitmq

import "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"

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

type RouteSelector interface {
	Select(key string) int
}

type Router struct {
	tags []string
	s    RouteSelector
	p    *DirectPublisher
}

func NewRouter(config DirectPublisherConfig, tags []string, s RouteSelector) Router {
	return Router{
		tags: tags,
		s:    s,
		p:    &DirectPublisher{Config: config},
	}
}

func (r *Router) Connect(conn *Connection) error {
	return r.p.Connect(conn)
}

func (r *Router) Close() error {
	return r.p.Close()
}

func (r *Router) Write(p []byte, key string) error {
	idx := r.s.Select(key)
	utils.Assert(idx < len(r.tags), "the index should be less that len(r.tags)")
	return r.p.Write(p, r.tags[idx])
}
