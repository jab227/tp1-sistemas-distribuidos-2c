package heap_test

import (
	"reflect"
	"testing"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/heap"
)

func TestHeap(t *testing.T) {
	h := heap.New[int]()
	values := []int{1, 4, 25, 49, 30, 21, 18, 87, 22}
	for _, v := range values {
		h.PushValue(v)
	}
	top5 := heap.TopN(h, 5)
	want := []int{87, 49, 30, 25, 22}

	if !reflect.DeepEqual(top5, want) {
		t.Fatalf("got %v, want %v", top5, want)
	}

}
