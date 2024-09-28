package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/lian_rr/keep/command"
	"github.com/lian_rr/keep/command/store/sqlite"
)

// ErrNotFound used when the searched element wasn't found.
var ErrNotFound = errors.New("not found")

// SqlOptFunc optional functions for Sql store.
type SqlOptFunc func(store *Sql) error

// Sql store
type Sql struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewSql returns a new SQL store.
func NewSql(logger *slog.Logger, opts ...SqlOptFunc) (Sql, error) {
	store := Sql{
		logger: logger,
	}

	for _, opt := range opts {
		if err := opt(&store); err != nil {
			return Sql{}, err
		}
	}

	if store.db == nil {
		return Sql{}, errors.New("missing db connection")
	}

	return store, nil
}

// Store stores a command on the sql store.
func (s *Sql) Store(ctx context.Context, cmd command.Command) error {
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

	_, err = tx.ExecContext(ctx, sqlite.InsertCommandQuery, cmd.ID.String(), cmd.Name, cmd.Description, cmd.Command)
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

		paramsQuery := fmt.Sprintf(sqlite.InsertParameterPartialQuery, strings.Join(placeholders, ","))
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

// ListCommands returns a list of all the commands
func (s *Sql) ListCommands(ctx context.Context) ([]command.Command, error) {
	rows, err := s.db.QueryContext(ctx, sqlite.GetAllCommandsQuery)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Warn("error closing rows when listing commands", slog.Any("error", err))
		}
	}()

	cmds := make([]command.Command, 0)
	for rows.Next() {
		var cmd command.Command
		if err := rows.Scan(&cmd.ID, &cmd.Name, &cmd.Description, &cmd.Command); err != nil {
			return nil, err
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// GetCommandByID returns a command. If the command doesn't exists, returns an ErrNotFound error.
func (s *Sql) GetCommandByID(ctx context.Context, id uuid.UUID) (command.Command, error) {
	row := s.db.QueryRowContext(ctx, sqlite.GetCommandbyIDQuery, id.String())
	var cmd command.Command
	if err := row.Scan(&cmd.ID, &cmd.Name, &cmd.Description, &cmd.Command); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return command.Command{}, ErrNotFound
		}
		return command.Command{}, err
	}

	rows, err := s.db.QueryContext(ctx, sqlite.GetParametersByCommandID, id.String())
	if err != nil {
		return command.Command{}, err
	}

	params := make([]command.Parameter, 0)
	for rows.Next() {
		var param command.Parameter
		if err := rows.Scan(&param.ID, &param.Name, &param.Description, &param.DefaultValue); err != nil {
			return command.Command{}, err
		}

		params = append(params, param)
	}

	cmd.Params = params
	return cmd, nil
}

// Close closes the db driver.
func (s *Sql) Close() error {
	return s.db.Close()
}

// WithSqliteDriver returns a sqlOptFunc that sets the config necessary for a SQLite store.
func WithSqliteDriver(ctx context.Context, path string) SqlOptFunc {
	return func(store *Sql) error {
		if store.db != nil {
			return errors.New("sql store already set")
		}

		db, err := sql.Open("sqlite3", fmt.Sprintf("%s/keep.db", path))
		if err != nil {
			return fmt.Errorf("error opening sqlite db: %v", err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}

		defer func() {
			err := tx.Rollback()
			if !errors.Is(err, sql.ErrTxDone) {
				store.logger.Warn("error rolling back init sqlite db config", slog.Any("error", err))
			}
		}()

		queries := []string{
			sqlite.CommandTableQuery,
			sqlite.ParametersTableQuery,
			sqlite.TagsTableQuery,
			sqlite.TagsAndCommandsTableQuery,
		}

		for _, query := range queries {
			if _, err := tx.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("error executing the query `%s`: %w", query, err)
			}
			store.logger.Debug("query executed successfully", slog.String("query", query))
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("error commiting transaction: %w", err)
		}

		store.logger.Debug("sqlite store initiatied successfully")

		store.db = db
		return nil
	}
}
