package command

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type store interface {
	Save(context.Context, Command) error
	GetCommandByID(context.Context, uuid.UUID) (Command, error)
	SearchCommand(context.Context, string) ([]Command, error)
	ListCommands(context.Context) ([]Command, error)
}

// Manager handles the command admin operations.
type Manager struct {
	store store
}

// NewManager returns a new Manager.
func NewManager(store store) (*Manager, error) {
	if store == nil {
		return nil, errors.New("nil store")
	}
	return &Manager{
		store: store,
	}, nil
}

// Add creates, saves and returns a new command validated.
func (m *Manager) Add(ctx context.Context, cmd Command) (Command, error) {
	var err error
	cmd.ID, err = uuid.NewV7()
	if err != nil {
		return Command{}, nil
	}

	if err := m.store.Save(ctx, cmd); err != nil {
		return Command{}, err
	}

	return cmd, nil
}

// GetCommand returns a command by ID.
func (m *Manager) GetOne(ctx context.Context, rawID string) (Command, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return Command{}, err
	}

	cmd, err := m.store.GetCommandByID(ctx, id)
	if err != nil {
		return Command{}, err
	}

	return cmd, nil
}

// SearchCommand returns a list of commands with a matching term.
func (m *Manager) Search(ctx context.Context, term string) ([]Command, error) {
	commands, err := m.store.SearchCommand(ctx, term)
	if err != nil {
		return nil, err
	}

	return commands, nil
}

// GetAll returns a list with all the commands.
func (m *Manager) GetAll(ctx context.Context) ([]Command, error) {
	commands, err := m.store.ListCommands(ctx)
	if err != nil {
		return nil, err
	}

	return commands, nil
}
