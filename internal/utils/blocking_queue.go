package utils

type BlockingQueue[T any] struct {
	channel chan T
}

func NewBlockingQueue[T any](capacity int) *BlockingQueue[T] {
	channel := make(chan T, capacity)
	return &BlockingQueue[T]{channel: channel}
}

func (bq *BlockingQueue[T]) Push(data T) {
	bq.channel <- data
}

func (bq *BlockingQueue[T]) Pop() (T, bool) {
	data, isClosed := <-bq.channel
	return data, isClosed
}

func (bq *BlockingQueue[T]) Close() {
	close(bq.channel)
}

func (bq *BlockingQueue[T]) Iter() <-chan T {
	return bq.channel
}
