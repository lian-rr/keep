package tui

import (
	"bytes"
	"log/slog"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/lian_rr/keep/command"
)

const (
	chromaLang      = "bash"
	chromaFormatter = "terminal16m"
	chromaStyle     = "catppuccin-frappe"
)

var paramHeaders = []string{
	"name",
	"description",
	"value",
}

type detailCommand struct {
	view        viewport.Model
	paramsTable *table.Table
	logger      *slog.Logger
}

func newDetailCommand(logger *slog.Logger) detailCommand {
	capitalizeHeaders := func(data []string) []string {
		for i := range data {
			data[i] = strings.ToUpper(data[i])
		}
		return data
	}

	tb := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(capitalizeHeaders(paramHeaders)...)

	return detailCommand{
		paramsTable: tb,
		view:        viewport.New(0, 0),
		logger:      logger,
	}
}

func (dc *detailCommand) SetContent(cmd command.Command) error {
	var b bytes.Buffer
	err := quick.Highlight(&b, cmd.Command, chromaLang, chromaFormatter, chromaStyle)
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(cmd.Params))
	for _, param := range cmd.Params {
		rows = append(rows, []string{param.Name, param.Description, param.DefaultValue})
	}

	dc.paramsTable.Data(table.NewStringData(rows...))

	content := containerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			cmd.Name,
			cmd.Description,
			b.String(),
			lipgloss.NewStyle().
				AlignHorizontal(lipgloss.Center).
				Render(dc.paramsTable.Render()),
		),
	)

	dc.view.SetContent(content)
	return nil
}

func (dc *detailCommand) View() string {
	return dc.view.View()
}

func (dc *detailCommand) SetSize(width, height int) {
	dc.view.Width, dc.view.Height = width, height
	w, _ := relativeDimensions(width, height, .50, .50)
	dc.paramsTable.Width(w)
}
