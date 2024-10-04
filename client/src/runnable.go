package src

type Runnable interface {
	Run(join chan error)
}
