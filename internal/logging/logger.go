package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const LoggerLevelEnvVariable = "LOGGER_LEVEL"

var slogLevelMap = map[string]slog.Level{
	"info":  slog.LevelInfo,
	"debug": slog.LevelDebug,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func InitLogger(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}

func InitLoggerWithString(level string) error {
	slogLevel, ok := slogLevelMap[strings.ToLower(level)]
	if !ok {
		return fmt.Errorf("invalid log level: %s", level)
	}

	InitLogger(slogLevel)
	return nil
}

func InitLoggerWithEnv() error {
	value, ok := os.LookupEnv(LoggerLevelEnvVariable)
	if !ok {
		return fmt.Errorf("environment variable %s not found", LoggerLevelEnvVariable)
	}

	level, ok := slogLevelMap[value]
	if !ok {
		return fmt.Errorf("invalid logger level %s", value)
	}

	InitLogger(level)
	return nil
}
