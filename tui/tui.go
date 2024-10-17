package tui

import (
	"context"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

// Tui contains the TUI logic.
type Tui struct {
	program *tea.Program
}

// New returns a new TUI container.
func New(ctx context.Context, logger *slog.Logger) (Tui, error) {
	model, err := newMainModel(logger)
	if err != nil {
		return Tui{}, fmt.Errorf("error starting the main model: %w", err)
	}

	return Tui{
		program: tea.NewProgram(
			model,
			tea.WithContext(ctx),
			tea.WithAltScreen(),
		),
	}, nil
}

// Start start the TUI app.
func (t *Tui) Start() error {
	if _, err := t.program.Run(); err != nil {
		return fmt.Errorf("error starting the TUI program: %w", err)
	}

	return nil
}
