package utils

import "fmt"

// Not goroutine-safe
var enabled bool = true

func Assert(condition bool, msg string) {
	if enabled && !condition {
		panic(msg)
	}
}

func Assertf(condition bool, format string, args ...any) {
	if enabled && !condition {
		msg := fmt.Sprintf(format, args...)
		panic(msg)
	}
}

func DisableAssertions() {
	enabled = false
}

func EnableAssertions() {
	enabled = true
}
