package tui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lian_rr/keep/command"
)

const title = "KEEP"

type mode int

const (
	_ = iota
	navigation
	create
	edit
	search
)

type mainModel struct {
	keys   keyMap
	help   help.Model
	logger *slog.Logger

	// panels
	commands   commandList
	detailView viewport.Model

	// styles
	titleStyle lipgloss.Style
}

func newMainModel(logger *slog.Logger) (*mainModel, error) {
	style := borderStyle.Align(lipgloss.Center, lipgloss.Center)
	vp := viewport.New(0, 0)
	vp.Style = style

	cmds := []command.Command{
		{
			Name:        "Run docker",
			Description: "Used for running docker",
		},
		{
			Name:        "Print env",
			Description: "Used for printing a desired env",
		},
	}

	model := mainModel{
		titleStyle: titleStyle,
		commands:   newCommandList("Commands", cmds),
		detailView: vp,
		keys:       defaultKeyMap,
		help:       help.New(),
		logger:     logger,
	}

	return &model, nil
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// key input
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.detailView, cmd = m.detailView.Update(msg)
			m.commands, cmd = m.commands.Update(msg)
			return m, cmd
		}
	// window resize
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.updateComponentsDimensions(msg.Width-h, msg.Height-v)
		return m, nil
	default:
		return m, nil
	}
}

func (m *mainModel) View() string {
	return docStyle.Render(borderStyle.Render(lipgloss.JoinVertical(lipgloss.Top,
		borderStyle.Render(m.titleStyle.Render(title)),
		borderStyle.Render(lipgloss.JoinHorizontal(
			lipgloss.Left,
			borderStyle.Render(m.commands.View()),
			// containerStyle.Render(m.detailView.View()),
		)),
		helpStyle.Render(m.help.View(m.keys)),
	)))
}

func (m *mainModel) Init() tea.Cmd {
	tea.SetWindowTitle(title)
	return nil
}

func (m *mainModel) updateComponentsDimensions(width, height int) {
	w, _ := relativeDimensions(width, 0, .985, 0)
	// title
	m.titleStyle = m.titleStyle.Width(w)

	// help
	m.help.Width = w

	// detail viewport
	w, h := relativeDimensions(width, height, .45, .80)
	m.detailView.Width, m.detailView.Height = w, h

	m.commands.SetSize(w, h)
}

func relativeDimensions(w, h int, pw, ph float32) (width, height int) {
	return int(float32(w) * pw), int(float32(h) * ph)
}
