package leader

type State uint8

const (
	StateLeading State = iota
	StateElecting
	StateMonitoring
)


