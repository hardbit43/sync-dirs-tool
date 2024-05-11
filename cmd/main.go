package main

import (
	"context"
	"finalTask/pkg/logger"
	"finalTask/pkg/synchronizer"
	"log/slog"
	"os"
	"time"
)

func main() {
	logger.InitLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer slog.Log(ctx, slog.LevelInfo.Level(), "synchronized")
	err := synchronizer.SyncDirs(ctx, synchronizer.SrcDir, synchronizer.DestDir)
	if err != nil {
		slog.Error("not synchronized", err)
		os.Exit(1)
	}
	time.Sleep(1 * time.Second)
}
