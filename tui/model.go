package tui

import (
	"context"
	"log/slog"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const title = "KEEP"

type model struct {
	ctx            context.Context
	commandManager manager

	keys   keyMap
	logger *slog.Logger

	// panels
	commands   listView
	detailView detailsView
	help       help.Model

	currentMode mode

	// styles
	titleStyle lipgloss.Style
}

func newModel(ctx context.Context, manager manager, logger *slog.Logger) (*model, error) {
	cmds, err := manager.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	detail := newDetailsView(logger)
	if len(cmds) > 0 {
		cmd, err := manager.GetOne(ctx, cmds[0].ID.String())
		if err != nil {
			return nil, err
		}
		cmds[0] = cmd
		detail.SetContent(cmds[0])
	}

	model := model{
		ctx:            ctx,
		commandManager: manager,
		titleStyle:     titleStyle,
		keys:           defaultKeyMap,
		commands:       newListView("Commands", cmds),
		detailView:     detail,
		help:           help.New(),
		currentMode:    navigationMode,
		logger:         logger,
	}

	return &model, nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// key input
	case tea.KeyMsg:
		// exit the app
		if key.Matches(msg, m.keys.forceQuit) {
			return m, tea.Quit
		}
		return m, m.inputRouter(msg)
	// window resize
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.updateComponentsDimensions(msg.Width-h, msg.Height-v)
		return m, nil
	// mode update
	case updateModeMsg:
		msg.updateMode(m)
		return m, nil
	}

	return m, nil
}

func (m *model) View() string {
	return docStyle.Render(
		borderStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				m.titleStyle.Render(title),
				borderStyle.Render(lipgloss.JoinHorizontal(
					lipgloss.Left,
					m.commands.View(),
					m.detailView.View(),
				)),
				helpStyle.Render(m.help.View(m.keys)),
			)))
}

func (m *model) Init() tea.Cmd {
	tea.SetWindowTitle(title)
	return nil
}

func (m *model) updateComponentsDimensions(width, height int) {
	m.logger.Debug("updating components dimensions")
	w, _ := relativeDimensions(width, 0, .985, 0)
	// title
	m.titleStyle = m.titleStyle.Width(w)

	// help
	m.help.Width = w

	w, h := relativeDimensions(width, height, .33, .90)
	// command explorer
	m.commands.SetSize(w, h)

	// detail view
	w, h = relativeDimensions(width, height, .80, .90)
	m.detailView.SetSize(w, h)
}

func (m *model) inputRouter(msg tea.KeyMsg) tea.Cmd {
	switch m.currentMode {
	case searchMode:
	case createMode:
	case editMode:
	case detailMode:
		return m.handleDetailInput(msg)
	default:
		return m.handleNavigationInput(msg)
	}

	return nil
}

func (m *model) handleNavigationInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.quit):
		return tea.Quit
	case key.Matches(msg, m.keys.enter):
		command, err := m.commands.selectedItem()
		if err != nil {
			m.logger.Error("error getting selected command", slog.Any("error", err))
			break
		}

		return changeMode(detailMode, func(m *model) {
			err := m.detailView.SetContent(*command.cmd)
			if err != nil {
				m.logger.Error("error setting detail view content", slog.Any("error", err))
			}
		})
	default:
		m.commands, cmd = m.commands.Update(msg)
		command, err := m.commands.selectedItem()
		if err != nil {
			m.logger.Error("error getting selected item", slog.Any("error", err))
			break
		}

		if !command.loaded {
			c, err := m.commandManager.GetOne(m.ctx, command.cmd.ID.String())
			if err != nil {
				m.logger.Error("error fetching command details", slog.Any("error", err))
				break
			}

			command.cmd.Params = c.Params
			command.loaded = true

			m.logger.Debug("command details fetched successfully", slog.Any("command", c))
		}
		m.detailView.SetContent(*command.cmd)
	}
	return cmd
}

func (m *model) handleDetailInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.back):
		return changeMode(navigationMode, nil)
	default:
		// NOTE: nothing for the moment
	}
	return cmd
}

func relativeDimensions(w, h int, pw, ph float32) (width, height int) {
	return int(float32(w) * pw), int(float32(h) * ph)
}
