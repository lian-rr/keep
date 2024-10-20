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
func NewSql(logger *slog.Logger, opts ...SqlOptFunc) (*Sql, error) {
	store := &Sql{
		logger: logger,
	}

	for _, opt := range opts {
		if err := opt(store); err != nil {
			return nil, err
		}
	}

	if store.db == nil {
		return nil, errors.New("missing db connection")
	}

	return store, nil
}

// Save stores a command on the sql store.
func (s *Sql) Save(ctx context.Context, cmd command.Command) error {
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

// SearchCommand returns a list of the commands with the matching term.
func (s *Sql) SearchCommand(ctx context.Context, term string) ([]command.Command, error) {
	rows, err := s.db.QueryContext(ctx, sqlite.SearchCommandQuery, term)
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
			sqlite.SearchTableQuery,
			sqlite.InsertCommandFtsTrigger,
			sqlite.UpdateCommandFtsTrigger,
			sqlite.DeleteCommandFtsTrigger,
			// testCommands,
			// testParams,
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

var testCommands = `
	INSERT INTO commands (id, name, description, command) VALUES 
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a5', 'CreateUser', 'Creates a new user in the system', 'useradd {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a6', 'DeleteUser', 'Removes a user from the system', 'userdel {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a7', 'UpdateUser', 'Updates user information', 'usermod -c "{{.info}}" {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a8', 'ListUsers', 'Retrieves a list of all users', 'cat /etc/passwd | grep {{.filter}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537a9', 'GetUser', 'Fetches details of a specific user', 'getent passwd {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537aa', 'ChangePassword', 'Updates the password for a user', 'passwd {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537ab', 'LockUser', 'Locks a user account to prevent access', 'usermod -L {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537ac', 'UnlockUser', 'Unlocks a previously locked user account', 'usermod -U {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537ad', 'GetLogs', 'Retrieves system logs for audit purposes', 'journalctl -xe --user {{.username}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537ae', 'SystemStatus', 'Checks the current status of the system', 'systemctl status {{.service}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537af', 'BackupDatabase', 'Backs up the specified database', 'pg_dump {{.database}} -f {{.output}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b0', 'RestoreDatabase', 'Restores the specified database from a backup', 'psql {{.database}} < {{.backup}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b1', 'CheckDiskSpace', 'Checks the disk space usage', 'df -h {{.path}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b2', 'StartService', 'Starts a specified service', 'systemctl start {{.service}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b3', 'StopService', 'Stops a specified service', 'systemctl stop {{.service}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b4', 'CheckServiceStatus', 'Checks the status of a specified service', 'systemctl status {{.service}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b5', 'CreateDirectory', 'Creates a new directory', 'mkdir -p {{.path}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b6', 'DeleteFile', 'Deletes a specified file', 'rm -f {{.filename}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b7', 'MoveFile', 'Moves a file to a new location', 'mv {{.source}} {{.destination}}'),
	('7f5f4b38-59ef-7e3c-8d6d-73e60c9537b8', 'CopyFile', 'Copies a file to a new location', 'cp {{.source}} {{.destination}}')
`

var testParams = `
	INSERT INTO parameters (id, command, name, description, value) VALUES 
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537a5', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a5', 'username', 'The name of the user to create', 'newuser'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537a6', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a6', 'username', 'The name of the user to delete', 'olduser'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537a7', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a7', 'info', 'New information for the user', 'Updated info'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537a8', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a8', 'filter', 'Pattern to filter the user list', '*'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537a9', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537a9', 'username', 'The name of the user to fetch', 'specificuser'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537aa', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537aa', 'username', 'The name of the user to change password', 'changeme'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537ab', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537ab', 'username', 'The name of the user to lock', 'lockeduser'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537ac', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537ac', 'username', 'The name of the user to unlock', 'unlockeduser'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537ad', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537ad', 'username', 'The name of the user to get logs for', 'loggeruser'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537ae', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537ae', 'service', 'The name of the service to check status', 'myservice'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537af', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537af', 'database', 'The name of the database to back up', 'mydatabase'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537b0', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b0', 'database', 'The name of the database to restore', 'mydatabase'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537b1', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b1', 'path', 'The path to check disk space', '/'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537b2', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b2', 'service', 'The name of the service to start', 'myservice'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537b3', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b3', 'service', 'The name of the service to stop', 'myservice'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537b4', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b4', 'service', 'The name of the service to check', 'myservice'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537b5', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b5', 'path', 'The path for the new directory', '/new/directory'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537b6', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b6', 'filename', 'The name of the file to delete', 'file.txt'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537b7', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b7', 'source', 'The source file to move', 'file.txt'),
	('1c3a5500-59ef-7e3c-8d6d-73e60c9537b8', '7f5f4b38-59ef-7e3c-8d6d-73e60c9537b8', 'source', 'The source file to copy', 'file.txt')
`
