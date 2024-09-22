package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/lian_rr/keep/command"
)

// Manager handles the command admin operations.
type Manager struct{}

// New returns a new Manager.
func New() Manager {
	return Manager{}
}

// Add creates, saves and returns a new command validated.
func Add(ctx context.Context) (command.Command, error) {
	return command.Command{}, nil
}

// Update updates a command and saves it.
func Update(ctx context.Context, id uuid.UUID) (command.Command, error) {
	return command.Command{}, nil
}

// Delete removes a command.
func Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}
