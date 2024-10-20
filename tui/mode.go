package tui

import tea "github.com/charmbracelet/bubbletea"

type mode int

const (
	_ = iota
	navigationMode
	detailMode
	createMode
	editMode
	searchMode
)

type updateModeMsg struct {
	updateMode func(*model)
}

func changeMode(newMode mode, handler func(*model)) tea.Cmd {
	return func() tea.Msg {
		return updateModeMsg{
			updateMode: func(m *model) {
				m.currentMode = newMode
				if handler != nil {
					handler(m)
				}
			},
		}
	}
}
