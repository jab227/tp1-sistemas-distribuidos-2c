package routing

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"github.com/rabbitmq/amqp091-go"
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

type Delivery = amqp091.Delivery

func (r *Router) Read() Delivery {
	delivery := r.m.Read()
	return delivery
}

func (r *Router) Write(p []byte, key string) error {
	idx := r.s.Select(key)
	utils.Assert(idx < len(r.tags), "the index should be less that len(r.tags)")
	return r.m.Write(p, r.tags[idx])
}

func (r *Router) Destroy() {
	r.m.Close()
}