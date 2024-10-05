package utils

type Thread[T Runnable] struct {
	runnable T
	join     chan error
}

func NewThread[T Runnable](runnable T) *Thread[T] {
	capacity := 1
	join := make(chan error, capacity)
	return &Thread[T]{runnable: runnable, join: join}
}

func (t *Thread[T]) Run() {
	go t.runnable.Run(t.join)
}

func (t *Thread[T]) Join() error {
	return <-t.join
}
