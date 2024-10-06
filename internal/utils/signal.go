package utils

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func MakeSignalHandler() <-chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	return signalChannel
}

func BlockUntilSignal(signal <-chan os.Signal, done <-chan struct{}, cancel context.CancelFunc) {
	select {
	case s := <-signal:
		slog.Info("exit", "signal", s)
		cancel()
		<-done
	case <-done:
		slog.Info("exit received done")
		cancel()
		return
	}
}
