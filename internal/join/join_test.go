package join_test

import (
	"cmp"
	"reflect"
	"slices"
	"testing"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/join"
)

type A struct {
	ID   string
	Data string
}

func (a A) GetID() string {
	return a.ID
}

type B struct {
	ID   string
	Data string
}

func (b B) GetID() string {
	return b.ID
}

func TestJoin(t *testing.T) {
	as := []A{{"1", "data1"}, {"2", "data2"}, {"3", "data3"}, {"3", "data9"}}
	bs := []B{{"1", "data4"}, {"6", "data5"}, {"5", "data6"}, {"3", "data7"}}

	firstTuple := join.Tuple[A, B]{A{"1", "data1"}, B{"1", "data4"}}
	secondTuple := join.Tuple[A, B]{A{"3", "data3"}, B{"3", "data7"}}
	thirdTuple := join.Tuple[A, B]{A{"3", "data9"}, B{"3", "data7"}}
	want := []join.Tuple[A, B]{firstTuple, secondTuple, thirdTuple}
	got := join.Join(as, bs)
	slices.SortFunc(got, func(a, b join.Tuple[A, B]) int {
		return cmp.Compare(a.X.ID, b.X.ID)
	})

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
