package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/lian_rr/keep/command"
)

// LocalOpts setup the optional Local configs.
type LocalOpts func(*Local) error

// ErrInvalidDir indicates that the root directory for the store is not valid.
var ErrInvalidDir = errors.New("invalid base directory")

// Local store
type Local struct {
	path   string
	db     *sql.DB
	logger *slog.Logger
}

// NewLocal returns a new Local store.
func NewLocal(ctx context.Context, logger *slog.Logger, opts ...LocalOpts) (store Local, shutdown func() error, err error) {
	store = Local{
		logger: logger,
	}

	for _, opt := range opts {
		if err := opt(&store); err != nil {
			return Local{}, nil, err
		}
	}

	if store.path == "" {
		path, err := os.UserHomeDir()
		if err != nil {
			return Local{}, nil, err
		}
		store.path = path
	}
	store.path = fmt.Sprintf("%s/.keep", store.path)

	if err := store.init(ctx); err != nil {
		return Local{}, nil, err
	}

	return store, store.db.Close, nil
}

func (s *Local) init(ctx context.Context) error {
	// TODO: move this logic to it's own package/logic if more data needs to be stored.
	err := os.Mkdir(s.path, 0o740)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("%w: %w", ErrInvalidDir, err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s/keep.db", s.path))
	if err != nil {
		return err
	}
	s.db = db

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		err := tx.Rollback()
		if !errors.Is(err, sql.ErrTxDone) {
			s.logger.Warn("error rolling back init local db config", slog.Any("error", err))
		}
	}()

	queries := []string{
		commandTableQuery,
		parametersTableQuery,
		tagsTableQuery,
		tagsAndCommandsTableQuery,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("error executing the query `%s`: %w", query, err)
		}
		s.logger.Debug("query executed successfully", slog.String("query", query))
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction: %w", err)
	}

	s.logger.Debug("Local store initiatied successfully")
	return nil
}

// Store stores a command on the local store.
func (s Local) Store(ctx context.Context, cmd command.Command) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		err := tx.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.logger.Warn("error rolling back store command", slog.Any("error", err))
			return
		}
	}()

	_, err = tx.ExecContext(ctx, insertCommandQuery, cmd.ID.String(), cmd.Name, cmd.Description, cmd.Command)
	if err != nil {
		return fmt.Errorf("error storing command: %w", err)
	}

	if len(cmd.Params) > 0 {
		placeholders := make([]string, 0, len(cmd.Params))
		args := make([]any, 0, len(cmd.Params)*5) // cap: number of params * attrs to store

		for _, param := range cmd.Params {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
			args = append(args, param.ID.String(), cmd.ID.String(), param.Name, param.Description, param.DefaultValue)
		}

		paramsQuery := fmt.Sprintf(insertParameterPartialQuery, strings.Join(placeholders, ","))
		_, err = tx.ExecContext(ctx, paramsQuery, args...)
		if err != nil {
			return fmt.Errorf("error storing parameters: %w", err)
		}
	}

	// TODO: add creating or getting the tags and setting the relationship

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction: %w", err)
	}

	s.logger.Debug("command stored successfully")
	return nil
}

// WithPath sets the optional path to the store
func WithPath(path string) LocalOpts {
	return func(store *Local) error {
		store.path = path
		store.logger.Debug("default local store path replaced", slog.String("path", path))
		return nil
	}
}
