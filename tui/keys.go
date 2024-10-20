package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	search    key.Binding
	quit      key.Binding
	forceQuit key.Binding
	enter     key.Binding
	back      key.Binding
}

var defaultKeyMap = keyMap{
	search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search")),
	quit: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc", "quit")),
	forceQuit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "force exit")),
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "enter")),
	back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back")),
}

func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.back,
		km.search,
		km.enter,
	}
}

func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
