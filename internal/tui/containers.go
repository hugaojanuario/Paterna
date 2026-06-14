package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hugaojanuario/Paterna/internal/container"
)

type containersLoadedMsg struct {
	rows []container.ContainerInfo
}

type errMsg struct {
	err error
}

type actionDoneMsg struct {
	text string
}

type ContainersModel struct {
	table  table.Model
	rows   []container.ContainerInfo
	status string
	err    error
}

func NewContainersModel() ContainersModel {
	columns := []table.Column{
		{Title: "ID", Width: 14},
		{Title: "NAME", Width: 28},
		{Title: "IMAGE", Width: 32},
		{Title: "STATUS", Width: 28},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	style.Selected = style.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(style)

	return ContainersModel{table: t}
}

func (m ContainersModel) Init() tea.Cmd {
	return loadContainers
}

func (m ContainersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case containersLoadedMsg:
		m.rows = msg.rows
		m.err = nil
		m.table.SetRows(toTableRows(msg.rows))
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case actionDoneMsg:
		m.status = msg.text
		return m, loadContainers

	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			id := m.selectedID()
			if id == "" {
				return m, nil
			}
			m.status = "reiniciando " + shortID(id) + "..."
			return m, restartContainer(id)

		case "s":
			id := m.selectedID()
			if id == "" {
				return m, nil
			}
			m.status = "iniciando " + shortID(id) + "..."
			return m, startContainer(id)

		case "x":
			id := m.selectedID()
			if id == "" {
				return m, nil
			}
			m.status = "parando " + shortID(id) + "..."
			return m, stopContainer(id)

		case "u":
			m.status = "atualizando..."
			return m, loadContainers
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m ContainersModel) View() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Paterna — Containers")

	footer := lipgloss.NewStyle().
		Faint(true).
		Render("↑↓: navegar  s: start  x: stop  r: restart  u: atualizar  q: sair")

	status := ""
	if m.err != nil {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("erro: " + m.err.Error())
	} else if m.status != "" {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Render(m.status)
	}

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n", header, m.table.View(), status, footer)
}

func (m ContainersModel) selectedID() string {
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.rows) {
		return ""
	}
	return m.rows[cursor].ID
}

func toTableRows(rows []container.ContainerInfo) []table.Row {
	result := make([]table.Row, 0, len(rows))
	for _, c := range rows {
		result = append(result, table.Row{
			shortID(c.ID),
			c.Name,
			c.Image,
			c.Status,
		})
	}
	return result
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func loadContainers() tea.Msg {
	rows, err := container.List(true)
	if err != nil {
		return errMsg{err: err}
	}
	return containersLoadedMsg{rows: rows}
}

func startContainer(id string) tea.Cmd {
	return func() tea.Msg {
		if err := container.StartContainer(id); err != nil {
			return errMsg{err: err}
		}
		return actionDoneMsg{text: "iniciado " + shortID(id)}
	}
}

func stopContainer(id string) tea.Cmd {
	return func() tea.Msg {
		if err := container.StopContainer(id); err != nil {
			return errMsg{err: err}
		}
		return actionDoneMsg{text: "parado " + shortID(id)}
	}
}

func restartContainer(id string) tea.Cmd {
	return func() tea.Msg {
		if err := container.RestartContainer(id); err != nil {
			return errMsg{err: err}
		}
		return actionDoneMsg{text: "reiniciado " + shortID(id)}
	}
}
