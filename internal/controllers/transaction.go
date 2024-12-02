package controllers

type TXN uint32

const (
	TXNBatch TXN = iota
	TXNSet
	TXNEnd
)
