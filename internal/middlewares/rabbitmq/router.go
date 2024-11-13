package rabbitmq

import (
	"hash"
	"hash/fnv"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	//	"log/slog"
)

type RoundRobinRouter struct {
	size int
	pos  int
}

func NewRoundRobinRouter(size int) RoundRobinRouter {
	return RoundRobinRouter{
		size: size,
		pos:  0,
	}
}

func (r *RoundRobinRouter) Select(key string) int {
	utils.Assert(r.pos < r.size, "should be less")
	pos := r.pos
	r.pos = (r.pos + 1) % r.size
	return pos
}

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
	r.hasher.Reset()
	r.hasher.Write([]byte(key))
	return int(r.hasher.Sum64() % uint64(r.idCount))
}

type GameReviewRouter struct {
}

func (g GameReviewRouter) Select(key string) int {
	if strings.EqualFold(key, "game") {
		return 0
	} else if strings.EqualFold(key, "review") {
		return 1
	} else {
		utils.Assert(false, "unreachable")
	}
	return -1
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
	// slog.Debug("Index chosen from router", "index", idx, "key", key, "tag", r.tags[idx])
	utils.Assert(idx < len(r.tags), "the index should be less that len(r.tags)")
	return r.p.Write(p, r.tags[idx])
}
