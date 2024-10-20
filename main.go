package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/lian_rr/keep/app"
	"github.com/lian_rr/keep/command"
	"github.com/lian_rr/keep/command/store"
	"github.com/lian_rr/keep/tui"
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

	logger, close, err := initLogger()
	if err != nil {
		return err
	}
	defer close()

	cfg, err := app.New(ctx, storePath, logger)
	if err != nil {
		return err
	}

	store, err := store.NewSql(logger, store.WithSqliteDriver(ctx, cfg.BasePath))
	if err != nil {
		slog.Error("error initializing the local store", slog.Any("error", err))
		return err
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Warn("error closing store", slog.Any("error", err))
			return
		}
		logger.Debug("Store closed successfully")
	}()

	manager, err := command.NewManager(store)
	if err != nil {
		slog.Error("error starting command manager", slog.Any("error", err))
		return err
	}

	ui, err := tui.New(ctx, manager, logger)
	if err != nil {
		return err
	}

	if err := ui.Start(); err != nil {
		return err
	}

	return nil
}

func initLogger() (logger *slog.Logger, close func() error, err error) {
	logLevel := slog.LevelInfo
	// if _, ok := os.LookupEnv(debugLogger); ok {
	logLevel = slog.LevelDebug
	// }

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
