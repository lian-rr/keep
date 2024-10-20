package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestManager_Add(t *testing.T) {
	mockErr := errors.New("mock error")
	testCmd := Command{
		Name:        "test command",
		Description: "this is a test command",
		Command:     "echo 'hello world'",
	}

	commandMatcher := func(cmd Command) bool {
		return cmd.Name == testCmd.Name &&
			cmd.Command == testCmd.Command &&
			cmd.Description == testCmd.Description
	}

	tests := []struct {
		name           string
		expectedError  error
		input          Command
		setExpectation func(mock *mockStore, ctx context.Context)
	}{
		{
			name:          "store returned an error",
			expectedError: mockErr,
			input:         testCmd,
			setExpectation: func(testMock *mockStore, ctx context.Context) {
				testMock.On("Save", ctx, mock.MatchedBy(commandMatcher)).Return(mockErr)
			},
		},
		{
			name:  "happy path",
			input: testCmd,
			setExpectation: func(testMock *mockStore, ctx context.Context) {
				testMock.On("Save", ctx, mock.MatchedBy(commandMatcher)).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			ctx := context.Background()

			if tt.setExpectation != nil {
				tt.setExpectation(store, ctx)
			}

			manager := Manager{
				store: store,
			}

			cmd, err := manager.Add(ctx, tt.input)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error(), "error not the expected")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.NotEqual(t, uuid.Nil, cmd.ID, "cmd ID not set")
			assert.True(t, commandMatcher(cmd), "command not the expected")
		})
	}
}

func TestManager_GetOne(t *testing.T) {
	mockErr := errors.New("mock error")
	id, err := uuid.NewV7()
	require.NoError(t, err)

	testCmd := Command{
		ID:          id,
		Name:        "test command",
		Description: "this is a test command",
		Command:     "echo 'hello world'",
	}

	tests := []struct {
		name           string
		expectedError  error
		input          string
		setExpectation func(mock *mockStore, ctx context.Context)
	}{
		{
			name:          "store returned an error",
			expectedError: mockErr,
			input:         id.String(),
			setExpectation: func(testMock *mockStore, ctx context.Context) {
				testMock.On("GetCommandByID", ctx, id).Return(Command{}, mockErr)
			},
		},
		{
			name:  "happy path",
			input: id.String(),
			setExpectation: func(testMock *mockStore, ctx context.Context) {
				testMock.On("GetCommandByID", ctx, id).Return(testCmd, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			ctx := context.Background()

			if tt.setExpectation != nil {
				tt.setExpectation(store, ctx)
			}

			manager := Manager{
				store: store,
			}

			cmd, err := manager.GetOne(ctx, tt.input)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error(), "error not the expected")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, testCmd, cmd, "cmd no the expected")
		})
	}
}

func TestManager_Search(t *testing.T) {
	mockErr := errors.New("mock error")
	id, err := uuid.NewV7()
	require.NoError(t, err)
	id2, err := uuid.NewV7()
	require.NoError(t, err)
	id3, err := uuid.NewV7()
	require.NoError(t, err)

	testCmds := []Command{
		{
			ID:          id,
			Name:        "test command",
			Description: "this is a test command",
			Command:     "echo 'hello world'",
		},
		{
			ID:          id2,
			Name:        "test command 2",
			Description: "this is a test command 2",
			Command:     "echo 'hello world 2'",
		},
		{
			ID:          id3,
			Name:        "test command 3",
			Description: "this is a test command 3",
			Command:     "echo 'hello world 3'",
		},
	}

	tests := []struct {
		name           string
		expectedError  error
		input          string
		setExpectation func(mock *mockStore, ctx context.Context)
	}{
		{
			name:          "store returned an error",
			expectedError: mockErr,
			input:         "test",
			setExpectation: func(testMock *mockStore, ctx context.Context) {
				testMock.On("SearchCommand", ctx, "test").Return(nil, mockErr)
			},
		},
		{
			name:  "happy path",
			input: "test",
			setExpectation: func(testMock *mockStore, ctx context.Context) {
				testMock.On("SearchCommand", ctx, "test").Return(testCmds, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			ctx := context.Background()

			if tt.setExpectation != nil {
				tt.setExpectation(store, ctx)
			}

			manager := Manager{
				store: store,
			}

			cmds, err := manager.Search(ctx, tt.input)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error(), "error not the expected")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, testCmds, cmds, "cmd no the expected")
		})
	}
}

func TestManager_GetAll(t *testing.T) {
	mockErr := errors.New("mock error")
	id, err := uuid.NewV7()
	require.NoError(t, err)
	id2, err := uuid.NewV7()
	require.NoError(t, err)
	id3, err := uuid.NewV7()
	require.NoError(t, err)

	testCmds := []Command{
		{
			ID:          id,
			Name:        "test command",
			Description: "this is a test command",
			Command:     "echo 'hello world'",
		},
		{
			ID:          id2,
			Name:        "test command 2",
			Description: "this is a test command 2",
			Command:     "echo 'hello world 2'",
		},
		{
			ID:          id3,
			Name:        "test command 3",
			Description: "this is a test command 3",
			Command:     "echo 'hello world 3'",
		},
	}

	tests := []struct {
		name           string
		expectedError  error
		setExpectation func(mock *mockStore, ctx context.Context)
	}{
		{
			name:          "store returned an error",
			expectedError: mockErr,
			setExpectation: func(testMock *mockStore, ctx context.Context) {
				testMock.On("ListCommands", ctx).Return(nil, mockErr)
			},
		},
		{
			name: "happy path",
			setExpectation: func(testMock *mockStore, ctx context.Context) {
				testMock.On("ListCommands", ctx).Return(testCmds, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			ctx := context.Background()

			if tt.setExpectation != nil {
				tt.setExpectation(store, ctx)
			}

			manager := Manager{
				store: store,
			}

			cmds, err := manager.GetAll(ctx)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error(), "error not the expected")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Equal(t, testCmds, cmds, "cmd no the expected")
		})
	}
}

type mockStore struct {
	mock.Mock
}

var _ store = (*mockStore)(nil)

func (m *mockStore) Save(ctx context.Context, cmd Command) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *mockStore) GetCommandByID(ctx context.Context, id uuid.UUID) (Command, error) {
	args := m.Called(ctx, id)
	cmd := args.Get(0)
	if cmd == nil {
		return Command{}, args.Error(1)
	}
	return cmd.(Command), args.Error(1)
}

func (m *mockStore) SearchCommand(ctx context.Context, term string) ([]Command, error) {
	args := m.Called(ctx, term)
	cmds := args.Get(0)
	if cmds == nil {
		return nil, args.Error(1)
	}
	return cmds.([]Command), args.Error(1)
}

func (m *mockStore) ListCommands(ctx context.Context) ([]Command, error) {
	args := m.Called(ctx)
	cmds := args.Get(0)
	if cmds == nil {
		return nil, args.Error(1)
	}
	return cmds.([]Command), args.Error(1)
}
