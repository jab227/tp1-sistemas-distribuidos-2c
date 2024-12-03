package heap

import (
	"container/heap"
	"log/slog"
)

type Value struct {
	AppId string
	Name  string
	Count int
}

type Heap []Value

func (hr Heap) Len() int           { return len(hr) }
func (hr Heap) Less(i, j int) bool { return hr[i].Count > hr[j].Count }
func (hr Heap) Swap(i, j int)      { hr[i], hr[j] = hr[j], hr[i] }

func (h *Heap) Push(x any) {
	*h = append(*h, x.(Value))
}
func (h *Heap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
func (h *Heap) PushValue(x any) {
	heap.Push(h, x)
}

func NewHeap() *Heap {
	hr := &Heap{}
	heap.Init(hr)
	return hr
}

func (hg *Heap) TopN(n int) []Value {
	top := make([]Value, 0, n)
	nmin := min(n, hg.Len())
	for i := 0; i < nmin; i++ {
		value, _ := heap.Pop(hg).(Value)
		slog.Debug("top value", "value", value)
		top = append(top, value)
	}
	return top
}
