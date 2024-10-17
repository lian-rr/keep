package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	quit   key.Binding
	search key.Binding
}

var defaultKeyMap = keyMap{
	quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("esc", "exit")),
	search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search")),
}

func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.quit,
		km.search,
	}
}

func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
