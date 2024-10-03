package examples

import (
	"context"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

func main() {
	// args and env config goes here

	// Context handling and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	signal := utils.MakeSignalHandler()
	// The done channel shoul be owned by the middleware and have this shape
	done := make(chan struct{}, 1)
	// Here goes the middleware and custom code, the custom code
	// should have a Done() method that return a channel <-chan
	// struct{} that closes when the goroutine is safely closed (graceful shutdown)
	utils.BlockUntilSignal(signal, done, cancel)
}
