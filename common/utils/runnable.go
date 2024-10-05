package utils

type Runnable interface {
	Run(join chan error)
}
