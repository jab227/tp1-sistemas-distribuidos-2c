package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func MakeSignalHandler() <-chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	return signalChannel
}

