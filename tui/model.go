package tui

import (
	"log/slog"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type mainModel struct {
	detailView viewport.Model
	keys       keyMap
	help       help.Model
	logger     *slog.Logger
}

func (m mainModel) Init() tea.Cmd {
	tea.SetWindowTitle("keep")

	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// key input
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.detailView, cmd = m.detailView.Update(cmd)
			return m, cmd
		}
	// window resize
	case tea.WindowSizeMsg:
		m.updateComponentsDimensions(msg.Width, msg.Height)
		return m, nil
	default:
		return m, nil
	}
}

func (m mainModel) View() string {
	center := lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Padding(0, 2, 0)

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(
			lipgloss.Left, m.detailView.View(),
		),
		center.Render(m.help.View(m.keys)),
	)
}

func newMainModel(logger *slog.Logger) (mainModel, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return mainModel{}, err
	}

	style := borderStyle.Align(lipgloss.Center, lipgloss.Center)
	vp := viewport.New(0, 0)
	vp.Style = style

	model := mainModel{
		detailView: vp,
		keys:       defaultKeyMap,
		help:       help.New(),
		logger:     logger,
	}

	model.updateComponentsDimensions(width, height)
	return model, nil
}

func relativeDimensions(w, h int, pw, ph float32) (width, height int) {
	return int(float32(w) * pw), int(float32(h) * ph)
}

func (m *mainModel) updateComponentsDimensions(width, height int) {
	// help
	m.help.Width = width

	// viewport
	w, h := relativeDimensions(width, height, .99, .96)
	m.detailView.Width, m.detailView.Height = w, h
}
