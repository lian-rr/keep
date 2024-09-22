package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"

	"github.com/lian_rr/keep/command"
	"github.com/lian_rr/keep/command/store"
)

const (
	debugLogger  = "KEEP_DEBUG"
	storePathEnv = "KEEP_STORE_PATH"
)

func main() {
	// exit once
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var storePath string
	if path, ok := os.LookupEnv(storePathEnv); ok {
		storePath = path
	}

	var storeOpts []store.LocalOpts
	if storePath != "" {
		storeOpts = append(storeOpts, store.WithPath(storePath))
	}

	logger, close, err := initLogger()
	if err != nil {
		return err
	}
	defer close()

	store, cancel, err := store.NewLocal(ctx, logger, storeOpts...)
	if err != nil {
		slog.Error("error initializing the local store", slog.Any("error", err))
		return err
	}
	defer cancel()

	id, _ := uuid.NewV6()
	cmd, err := command.New(id, "test command", "command for testing 2", "echo '{{.text}} - {{.text2}}'")
	if err != nil {
		return nil
	}

	store.Store(ctx, cmd)

	return nil
}

func initLogger() (logger *slog.Logger, close func() error, err error) {
	logLevel := slog.LevelInfo
	if _, ok := os.LookupEnv(debugLogger); ok {
		logLevel = slog.LevelDebug
	}

	file, err := os.OpenFile("/tmp/keep.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, nil, err
	}

	logger = slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)
	return logger, file.Close, nil
}
