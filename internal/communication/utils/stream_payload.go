package utils

type StreamPayload struct {
	Data []byte
}

func (s *StreamPayload) Sizeof() int {
	return len(s.Data)
}

func (s *StreamPayload) Marshall() []byte {
	return s.Data
}

func (s *StreamPayload) Unmarshall(data []byte) {
	s.Data = data
}
