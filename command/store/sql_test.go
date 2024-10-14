package store

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lian_rr/keep/command"
	"github.com/lian_rr/keep/command/store/sqlite"
)

func TestNewLocal(t *testing.T) {
	tests := []struct {
		name           string
		driverFunc     func(db *sql.DB) SqlOptFunc
		expectedErrMsg string
	}{
		{
			name:           "missing driver",
			expectedErrMsg: "missing db connection",
			driverFunc: func(_ *sql.DB) SqlOptFunc {
				return func(_ *Sql) error {
					return nil
				}
			},
		},
		{
			name:           "error executing opts",
			expectedErrMsg: "mock error",
			driverFunc: func(_ *sql.DB) SqlOptFunc {
				return func(_ *Sql) error {
					return errors.New("mock error")
				}
			},
		},
		{
			name: "happy path",
			driverFunc: func(db *sql.DB) SqlOptFunc {
				return func(store *Sql) error {
					store.db = db
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			db, _, err := sqlmock.New()
			require.NoError(t, err)

			got, err := NewSql(logger, tt.driverFunc(db))
			if tt.expectedErrMsg != "" {
				assert.ErrorContains(t, err, tt.expectedErrMsg, "error message not the expected")
				return
			}

			assert.NoError(t, err, "err unexpected")
			assert.Equal(t, db, got.db, "db instance not the expected")
			assert.Equal(t, logger, got.logger, "logger instance not the expected")
		})
	}
}

func TestSql_Save(t *testing.T) {
	id, err := uuid.NewV7()
	require.NoError(t, err)

	paramID1, err := uuid.NewV7()
	require.NoError(t, err)

	paramID2, err := uuid.NewV7()
	require.NoError(t, err)

	cmd := command.Command{
		ID:          id,
		Name:        "cmd 1",
		Description: "command 1",
		Command:     "echo '{{text}} - {{text2}}'",
		Params: []command.Parameter{
			{
				ID:           paramID1,
				Name:         "text",
				Description:  "param 1",
				DefaultValue: "hello",
			},
			{
				ID:           paramID2,
				Name:         "text2",
				Description:  "param 2",
				DefaultValue: "bye",
			},
		},
	}
	mockErr := errors.New("mock error")

	tests := []struct {
		name             string
		cmd              command.Command
		setMockCallsFunc func(cmd command.Command, mock sqlmock.Sqlmock)
		validateLogs     func(buf bytes.Buffer)
		expectedErrMsg   string
	}{
		{
			name: "error starting the transaction",
			setMockCallsFunc: func(_ command.Command, mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(mockErr)
			},
			expectedErrMsg: "error starting transaction",
		},
		{
			name: "error inserting command",
			cmd:  cmd,
			setMockCallsFunc: func(cmd command.Command, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec(sqlite.InsertCommandQuery).
					WithArgs(cmd.ID, cmd.Name, cmd.Description, cmd.Command).
					WillReturnError(mockErr)

				mock.ExpectRollback()
			},
			expectedErrMsg: "error storing command",
		},
		{
			name: "error inserting command - failed rollback",
			cmd:  cmd,
			setMockCallsFunc: func(cmd command.Command, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec(sqlite.InsertCommandQuery).
					WithArgs(cmd.ID, cmd.Name, cmd.Description, cmd.Command).
					WillReturnError(mockErr)

				mock.ExpectRollback().WillReturnError(mockErr)
			},
			validateLogs: func(buf bytes.Buffer) {
				assert.NotEmpty(t, buf, "buffer empty")
				eMsg := `level=WARN msg="error rolling back store command" error="mock error"`
				assert.Equal(t, eMsg, strings.TrimSpace(buf.String()), "log not the expected")
			},
			expectedErrMsg: "error storing command",
		},
		{
			name: "error inserting params",
			cmd:  cmd,
			setMockCallsFunc: func(cmd command.Command, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec(sqlite.InsertCommandQuery).
					WithArgs(cmd.ID, cmd.Name, cmd.Description, cmd.Command).
					WillReturnResult(sqlmock.NewResult(1, 1))

				paramsValue := make([]driver.Value, 0, len(cmd.Params)*5)
				for _, p := range cmd.Params {
					paramsValue = append(paramsValue, p.ID.String(), cmd.ID.String(), p.Name, p.Description, p.DefaultValue)
				}

				mock.ExpectExec(fmt.Sprintf(sqlite.InsertParameterPartialQuery, "(?, ?, ?, ?, ?),(?, ?, ?, ?, ?)")).
					WithArgs(paramsValue...).
					WillReturnError(mockErr)
				mock.ExpectRollback()
			},
			expectedErrMsg: "error storing parameters",
		},
		{
			name: "error committing",
			cmd:  cmd,
			setMockCallsFunc: func(cmd command.Command, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec(sqlite.InsertCommandQuery).
					WithArgs(cmd.ID, cmd.Name, cmd.Description, cmd.Command).
					WillReturnResult(sqlmock.NewResult(1, 1))

				paramsValue := make([]driver.Value, 0, len(cmd.Params)*5)
				for _, p := range cmd.Params {
					paramsValue = append(paramsValue, p.ID.String(), cmd.ID.String(), p.Name, p.Description, p.DefaultValue)
				}

				mock.ExpectExec(fmt.Sprintf(sqlite.InsertParameterPartialQuery, "(?, ?, ?, ?, ?),(?, ?, ?, ?, ?)")).
					WithArgs(paramsValue...).
					WillReturnResult(sqlmock.NewResult(2, 2))

				mock.ExpectCommit().WillReturnError(mockErr)
			},
			expectedErrMsg: "error commiting transaction",
		},
		{
			name: "error committing",
			cmd:  cmd,
			setMockCallsFunc: func(cmd command.Command, mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec(sqlite.InsertCommandQuery).
					WithArgs(cmd.ID, cmd.Name, cmd.Description, cmd.Command).
					WillReturnResult(sqlmock.NewResult(1, 1))

				paramsValue := make([]driver.Value, 0, len(cmd.Params)*5)
				for _, p := range cmd.Params {
					paramsValue = append(paramsValue, p.ID.String(), cmd.ID.String(), p.Name, p.Description, p.DefaultValue)
				}

				mock.ExpectExec(fmt.Sprintf(sqlite.InsertParameterPartialQuery, "(?, ?, ?, ?, ?),(?, ?, ?, ?, ?)")).
					WithArgs(paramsValue...).
					WillReturnResult(sqlmock.NewResult(2, 2))

				mock.ExpectCommit()
			},
			validateLogs: func(buf bytes.Buffer) {
				assert.NotEmpty(t, buf, "buffer empty")
				eMsg := `level=DEBUG msg="command stored successfully"`
				assert.Equal(t, eMsg, strings.TrimSpace(buf.String()), "log not the expected")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)

			tt.setMockCallsFunc(tt.cmd, mock)

			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
				// remove the time key
				ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.Attr{}
					}
					return a
				},
			}))

			store := Sql{
				logger: logger,
				db:     db,
			}

			err = store.Save(context.Background(), tt.cmd)
			if tt.validateLogs != nil {
				tt.validateLogs(buf)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "expectations not met")
			if tt.expectedErrMsg != "" {
				assert.ErrorContains(t, err, tt.expectedErrMsg, "error not the expected")
				return
			}

			assert.NoError(t, err, "unexpected error")
		})
	}
}

func TestSql_ListCommands(t *testing.T) {
	mockErr := errors.New("mock err")

	id, err := uuid.NewV7()
	require.NoError(t, err)
	id2, err := uuid.NewV7()
	require.NoError(t, err)
	id3, err := uuid.NewV7()
	require.NoError(t, err)

	cmds := []command.Command{
		{
			ID:          id,
			Name:        "test command",
			Description: "command used for testing",
			Command:     "echo '{{text}} - {{text2}}'",
		},
		{
			ID:          id2,
			Name:        "kill process running on port",
			Description: "kills the process running on the provided port",
			Command:     "lsof -t -i:{{port}} | xargs kill",
		},
		{
			ID:          id3,
			Name:        "Squash the last N commits",
			Description: "Squash the last N number of commits",
			Command:     "git reset --soft HEAD~num_commits && git commit",
		},
	}

	tests := []struct {
		name             string
		expectedErrorMsg string
		setMockCalls     func(mock sqlmock.Sqlmock)
		expectedOut      []command.Command
	}{
		{
			name:             "unexpected error getting command",
			expectedErrorMsg: mockErr.Error(),
			setMockCalls: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlite.GetAllCommandsQuery).WillReturnError(mockErr)
			},
		},
		{
			name:        "command found",
			expectedOut: cmds,
			setMockCalls: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"uuid", "name", "description", "command"})
				for _, cmd := range cmds {
					rows.AddRow(cmd.ID, cmd.Name, cmd.Description, cmd.Command)
				}

				mock.ExpectQuery(sqlite.GetAllCommandsQuery).
					WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)

			if tt.setMockCalls != nil {
				tt.setMockCalls(mock)
			}

			store := Sql{
				db: db,
			}

			got, err := store.ListCommands(context.Background())

			assert.NoError(t, mock.ExpectationsWereMet(), "expectations not met")
			if tt.expectedErrorMsg != "" {
				assert.ErrorContains(t, err, tt.expectedErrorMsg, "error not the expected")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expectedOut, got, "command not the expected")
		})
	}
}

func TestSql_GetCommandByID(t *testing.T) {
	mockErr := errors.New("mock err")

	id, err := uuid.NewV7()
	require.NoError(t, err)
	paramID1, err := uuid.NewV7()
	require.NoError(t, err)
	paramID2, err := uuid.NewV7()
	require.NoError(t, err)

	cmd := command.Command{
		ID:          id,
		Name:        "test command",
		Description: "command used for testing",
		Command:     "echo '{{text}} - {{text2}}'",
		Params: []command.Parameter{
			{
				ID:           paramID1,
				Name:         "text",
				Description:  "text param 1",
				DefaultValue: "hello",
			},
			{
				ID:           paramID2,
				Name:         "text2",
				Description:  "text param 2",
				DefaultValue: "bye",
			},
		},
	}

	tests := []struct {
		name             string
		expectedErrorMsg string
		setMockCalls     func(mock sqlmock.Sqlmock)
		expectedOut      command.Command
	}{
		{
			name:             "unexpected error getting command",
			expectedErrorMsg: mockErr.Error(),
			setMockCalls: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlite.GetCommandbyIDQuery).WillReturnError(mockErr)
			},
		},
		{
			name:             "not found command",
			expectedErrorMsg: ErrNotFound.Error(),
			setMockCalls: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlite.GetCommandbyIDQuery).WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:             "error getting params",
			expectedErrorMsg: mockErr.Error(),
			setMockCalls: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"uuid", "name", "description", "command"}).
					AddRow(cmd.ID, cmd.Name, cmd.Description, cmd.Command)

				mock.ExpectQuery(sqlite.GetCommandbyIDQuery).
					WithArgs(id.String()).
					WillReturnRows(rows)

				mock.ExpectQuery(sqlite.GetParametersByCommandID).
					WithArgs(id.String()).
					WillReturnError(mockErr)
			},
		},
		{
			name:        "command found",
			expectedOut: cmd,
			setMockCalls: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"uuid", "name", "description", "command"}).
					AddRow(cmd.ID, cmd.Name, cmd.Description, cmd.Command)

				mock.ExpectQuery(sqlite.GetCommandbyIDQuery).
					WithArgs(id.String()).
					WillReturnRows(rows)

				rows = sqlmock.NewRows([]string{"uuid", "name", "description", "value"})
				for _, param := range cmd.Params {
					rows.AddRow(param.ID, param.Name, param.Description, param.DefaultValue)
				}

				mock.ExpectQuery(sqlite.GetParametersByCommandID).
					WithArgs(id.String()).
					WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)

			if tt.setMockCalls != nil {
				tt.setMockCalls(mock)
			}

			store := Sql{
				db: db,
			}

			got, err := store.GetCommandByID(context.Background(), id)

			assert.NoError(t, mock.ExpectationsWereMet(), "expectations not met")
			if tt.expectedErrorMsg != "" {
				assert.ErrorContains(t, err, tt.expectedErrorMsg, "error not the expected")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expectedOut, got, "command not the expected")
		})
	}
}

func TestSql_SearchCommand(t *testing.T) {
	mockErr := errors.New("mock err")

	id, err := uuid.NewV7()
	require.NoError(t, err)
	id2, err := uuid.NewV7()
	require.NoError(t, err)
	id3, err := uuid.NewV7()
	require.NoError(t, err)

	cmds := []command.Command{
		{
			ID:          id,
			Name:        "test command",
			Description: "command used for testing",
			Command:     "echo '{{text}} - {{text2}}'",
		},
		{
			ID:          id2,
			Name:        "kill process running on port",
			Description: "kills the process running on the provided port",
			Command:     "lsof -t -i:{{port}} | xargs kill",
		},
		{
			ID:          id3,
			Name:        "Squash the last N commits",
			Description: "Squash the last N number of commits",
			Command:     "git reset --soft HEAD~num_commits && git commit",
		},
	}

	tests := []struct {
		name             string
		expectedErrorMsg string
		searchTerm       string
		setMockCalls     func(mock sqlmock.Sqlmock)
		expectedOut      []command.Command
	}{
		{
			name:             "unexpected error getting command",
			searchTerm:       "test",
			expectedErrorMsg: mockErr.Error(),
			setMockCalls: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlite.SearchCommandQuery).WillReturnError(mockErr)
			},
		},
		{
			name:        "commands found",
			expectedOut: cmds,
			searchTerm:  "whatever",
			setMockCalls: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"uuid", "name", "description", "command"})
				for _, cmd := range cmds {
					rows.AddRow(cmd.ID, cmd.Name, cmd.Description, cmd.Command)
				}

				mock.ExpectQuery(sqlite.SearchCommandQuery).
					WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)

			if tt.setMockCalls != nil {
				tt.setMockCalls(mock)
			}

			store := Sql{
				db: db,
			}

			got, err := store.SearchCommand(context.Background(), tt.searchTerm)

			assert.NoError(t, mock.ExpectationsWereMet(), "expectations not met")
			if tt.expectedErrorMsg != "" {
				assert.ErrorContains(t, err, tt.expectedErrorMsg, "error not the expected")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, tt.expectedOut, got, "command not the expected")
		})
	}
}
