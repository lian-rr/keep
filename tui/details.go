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
	chromaLang      = "fish"
	chromaFormatter = "terminal16m"
	chromaStyle     = "catppuccin-frappe"
)

var paramHeaders = []string{
	"name",
	"description",
	"default value",
}

type detailsView struct {
	view        viewport.Model
	infoTable   *table.Table
	paramsTable *table.Table
	logger      *slog.Logger

	// styles
	titleStyle   lipgloss.Style
	contentStyle lipgloss.Style
}

func newDetailsView(logger *slog.Logger) detailsView {
	capitalizeHeaders := func(data []string) []string {
		for i := range data {
			data[i] = strings.ToUpper(data[i])
		}
		return data
	}

	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor))

	infoTable := table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle()

			if col != 0 {
				style = style.MarginLeft(1)
			}

			return style
		})

	params := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		Headers(capitalizeHeaders(paramHeaders)...)

	return detailsView{
		infoTable:   infoTable,
		paramsTable: params,
		view:        vp,
		logger:      logger,
		titleStyle:  titleStyle,
		contentStyle: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 8),
	}
}

func (dc *detailsView) SetContent(cmd command.Command) error {
	var b bytes.Buffer
	if err := quick.Highlight(&b, cmd.Command, chromaLang, chromaFormatter, chromaStyle); err != nil {
		return err
	}

	rows := make([][]string, 0, len(cmd.Params))
	for _, param := range cmd.Params {
		rows = append(rows, []string{param.Name, param.Description, param.DefaultValue})
	}

	dc.paramsTable.Data(table.NewStringData(rows...))

	dc.infoTable.Data(table.NewStringData([][]string{
		{labelStyle.Render("Title"), headerStyle.Render(cmd.Name)},
		{labelStyle.Render("Description"), headerStyle.Render(cmd.Description)},
		{labelStyle.Render("Command"), headerStyle.Render(b.String())},
	}...))

	style := lipgloss.NewStyle()
	content := dc.contentStyle.
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				dc.infoTable.Render(),
				style.MarginLeft(1).Render(labelStyle.Render("Parameters")),
				style.MarginLeft(2).Render(dc.paramsTable.Render()),
			),
		)

	dc.view.SetContent(content)
	return nil
}

func (dc *detailsView) View() string {
	return dc.view.View()
}

func (dc *detailsView) SetSize(width, height int) {
	dc.titleStyle.Width(width)

	dc.view.Width, dc.view.Height = width, height
	w, _ := relativeDimensions(width, height, .50, .50)

	dc.paramsTable.Width(w)
	dc.logger.Debug("size set")
}
