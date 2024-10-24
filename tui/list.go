package tui

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/lian_rr/keep/command"
)

type listView struct {
	list list.Model
}

func newListView(title string, commands []command.Command) listView {
	items := make([]list.Item, 0, len(commands))
	for _, cmd := range commands {
		items = append(items, &listItem{
			title: cmd.Name,
			desc:  cmd.Description,
			cmd:   &cmd,
		})
	}

	view := list.New(items, list.NewDefaultDelegate(), 0, 0)
	view.Title = title
	view.DisableQuitKeybindings()
	view.SetFilteringEnabled(false)
	view.SetShowHelp(false)

	return listView{
		list: view,
	}
}

func (l *listView) Update(msg tea.Msg) (listView, tea.Cmd) {
	var cmd tea.Cmd
	l.list, cmd = l.list.Update(msg)
	return *l, cmd
}

func (l listView) View() string {
	return containerStyle.Render(l.list.View())
}

func (l *listView) SetSize(w, h int) {
	l.list.SetSize(w, h)
}

func (l *listView) selectedItem() (*listItem, error) {
	command, ok := l.list.SelectedItem().(*listItem)
	if !ok {
		return nil, errors.New("invalid item selected")
	}

	return command, nil
}

type listItem struct {
	title  string
	desc   string
	cmd    *command.Command
	loaded bool
}

var _ list.Item = (*listItem)(nil)

func (i listItem) Title() string {
	return i.title
}

func (i listItem) Description() string {
	return i.desc
}

func (i listItem) FilterValue() string {
	return i.title
}
