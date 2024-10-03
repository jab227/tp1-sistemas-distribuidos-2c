package heap

import (
	"cmp"
	"container/heap"
)

type Heap[T cmp.Ordered] []T

func (h Heap[T]) Len() int           { return len(h) }
func (h Heap[T]) Less(i, j int) bool { return h[i] > h[j] }
func (h Heap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// DON'T USE
func (h *Heap[T]) Push(x any) {
	*h = append(*h, x.(T))
}

func (h *Heap[T]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func New[T cmp.Ordered]() *Heap[T] {
	h := &Heap[T]{}
	heap.Init(h)
	return h
}

func (h *Heap[T]) PushValue(x T) {
	heap.Push(h, x)
}

func (h *Heap[T]) PopValue() T {
	return heap.Pop(h).(T)
}

func TopN[T cmp.Ordered](h *Heap[T], n int) []T {
	top := make([]T, n)
	nmin := min(n, h.Len())
	for i := 0; i < nmin; i++ {
		top[i] = h.PopValue()
	}
	return top
}
