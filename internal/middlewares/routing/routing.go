package routing

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type RouteSelector interface {
	Select(key string) int
}

type Router struct {
	tags []string
	s    RouteSelector
	m    *client.IOManager
}

func NewRouter(input client.InputType, tags []string, selector RouteSelector) (Router, error) {
	manager := &client.IOManager{}
	if err := manager.Connect(input, client.DirectPublisher); err != nil {
		return Router{}, fmt.Errorf("couldn't create router: %w", err)
	}

	return Router{m: manager, tags: tags, s: selector}, nil
}

func (r *Router) Read(onRead func([]byte) error) error {
	delivery := r.m.Read()
	error := onRead(delivery.Body)
	// NOTE(Juan): should we acknowledge in case of err?
	delivery.Ack(false)
	return error
}

func (r *Router) Write(p []byte, key string) error {
	idx := r.s.Select(key)
	utils.Assert(idx < len(r.tags), "the index should be less that len(r.tags)")
	return r.m.Write(p, r.tags[idx])
}

func (r *Router) Destroy() {
	r.m.Close()
}
