package tui

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian_rr/keep/command"
)

type commandList struct {
	list list.Model
}

func newCommandList(title string, commands []command.Command) commandList {
	items := make([]list.Item, 0, len(commands))
	for _, cmd := range commands {
		items = append(items, &commandItem{
			title: cmd.Name,
			desc:  cmd.Description,
			cmd:   &cmd,
		})
	}

	listView := list.New(items, list.NewDefaultDelegate(), 0, 0)
	listView.Title = title
	listView.DisableQuitKeybindings()
	listView.SetFilteringEnabled(false)
	listView.SetShowHelp(false)

	return commandList{
		list: listView,
	}
}

func (l *commandList) Update(msg tea.Msg) (commandList, tea.Cmd) {
	var cmd tea.Cmd
	l.list, cmd = l.list.Update(msg)
	return *l, cmd
}

func (l commandList) View() string {
	return containerStyle.Render(l.list.View())
}

func (l *commandList) SetSize(w, h int) {
	l.list.SetSize(w, h)
}

func (l *commandList) selectedItem() (*commandItem, error) {
	command, ok := l.list.SelectedItem().(*commandItem)
	if !ok {
		return nil, errors.New("invalid item selected")
	}

	return command, nil
}

type commandItem struct {
	title  string
	desc   string
	cmd    *command.Command
	loaded bool
}

var _ list.Item = (*commandItem)(nil)

func (i commandItem) Title() string {
	return i.title
}

func (i commandItem) Description() string {
	return i.desc
}

func (i commandItem) FilterValue() string {
	return i.title
}
