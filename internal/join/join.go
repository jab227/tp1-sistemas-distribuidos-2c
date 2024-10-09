package join


type HashID interface {
	GetID() string
}

type Tuple[T, U any] struct {
	X T
	Y U
}

func Join[T, U HashID](x []T, y []U) []Tuple[T, U] {
	output := make([]Tuple[T, U], 0, min(len(x), len(y)))
	table := make(map[string][]T, len(x))
	for _, r := range x {
		table[r.GetID()] = append(table[r.GetID()], r)
	}
	for _, p := range y {
		id := p.GetID()
		ts, ok := table[id]
		if !ok {
			continue
		}
		for _, t := range ts {
			output = append(output, Tuple[T, U]{X: t, Y: p})
		}
	}

	return output
}
