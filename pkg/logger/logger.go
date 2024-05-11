package logger

import (
	"log/slog"
	"os"
)

func InitLogger() {
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("error init logger", err)
	}
	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(file, opts))
	slog.SetDefault(logger)
}
