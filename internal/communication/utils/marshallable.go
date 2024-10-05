package utils

type Marshallable interface {
	Marshall() []byte
	Unmarshall(data []byte)
}
