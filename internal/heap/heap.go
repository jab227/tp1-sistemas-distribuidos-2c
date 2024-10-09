package heap

import (
	"container/heap"
	"slices"

	models "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
)

type HeapGames []models.Game

func (h HeapGames) Len() int           { return len(h) }
func (h HeapGames) Less(i, j int) bool { return h[i].AvgPlayTime < h[j].AvgPlayTime }
func (h HeapGames) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// DON'T USE Push and Pop
func (h *HeapGames) Push(x any) {
	*h = append(*h, x.(models.Game))
}
func (h *HeapGames) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func NewHeapGames() *HeapGames {
	h := &HeapGames{}
	heap.Init(h)
	return h
}

func (h *HeapGames) PushValue(x models.Game) {
	heap.Push(h, x)
}

func (h *HeapGames) PopValue() models.Game {
	return heap.Pop(h).(models.Game)
}

func (h *HeapGames) TopNGames(n uint64) []models.Game {
	top := make([]models.Game, n)
	nmin := min(n, uint64(h.Len()))

	for i := 0; i < int(nmin); i++ {
		top[i] = h.PopValue()
	}
	slices.Reverse(top)
	return top
}

type HeapReviews []models.Review

func (hr HeapReviews) Len() int           { return len(hr) }
func (hr HeapReviews) Less(i, j int) bool { return hr[i].Score < hr[j].Score }
func (hr HeapReviews) Swap(i, j int)      { hr[i], hr[j] = hr[j], hr[i] }

// DON'T USE Push and Pop
func (h *HeapReviews) Push(x any) {
	*h = append(*h, x.(models.Review))
}
func (h *HeapReviews) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func NewHeapReviews() *HeapReviews {
	hr := &HeapReviews{}
	heap.Init(hr)
	return hr
}

func (h *HeapReviews) PushValue(x models.Review) {
	heap.Push(h, x)
}

func (h *HeapReviews) PopValue() models.Review {
	return heap.Pop(h).(models.Review)
}

func (hg *HeapReviews) TopNReviews(n int) []models.Review {
	top := make([]models.Review, n)
	nmin := min(n, hg.Len())

	for i := 0; i < nmin; i++ {
		topTmp, _ := hg.Pop().(models.Review)
		top[i] = topTmp
	}
	slices.Reverse(top)
	return top
}


