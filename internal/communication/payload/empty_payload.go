package payload

type Empty struct {
}

func NewEmpty() *Empty {
	return &Empty{}
}

func (s *Empty) Sizeof() int {
	return 0
}

func (s *Empty) Marshall() []byte {
	return make([]byte, 0)
}

func (s *Empty) Unmarshall(data []byte) {
}
