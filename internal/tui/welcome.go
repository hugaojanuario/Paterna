package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hugaojanuario/Paterna/internal/version"
)

const asciiPaterna = `██████╗  █████╗ ████████╗███████╗██████╗ ███╗   ██╗ █████╗
██╔══██╗██╔══██╗╚══██╔══╝██╔════╝██╔══██╗████╗  ██║██╔══██╗
██████╔╝███████║   ██║   █████╗  ██████╔╝██╔██╗ ██║███████║
██╔═══╝ ██╔══██║   ██║   ██╔══╝  ██╔══██╗██║╚██╗██║██╔══██║
██║     ██║  ██║   ██║   ███████╗██║  ██║██║ ╚████║██║  ██║
╚═╝     ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝`

type menuItem struct {
	label   string
	command string
	enabled bool
}

var menuItems = []menuItem{
	{label: "Containers", command: "paterna containers", enabled: true},
	{label: "Métricas", command: "paterna metrics", enabled: false},
	{label: "Alertas", command: "paterna alerts", enabled: false},
	{label: "Configurações", command: "paterna settings", enabled: false},
	{label: "Sair", command: "paterna exit", enabled: true},
}

type WelcomeModel struct {
	width  int
	height int
	cursor int
	status string
}

func NewWelcomeModel() WelcomeModel {
	return WelcomeModel{}
}

func (m WelcomeModel) Init() tea.Cmd {
	return nil
}

func (m WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			m.cursor = (m.cursor - 1 + len(menuItems)) % len(menuItems)
			m.status = ""
			return m, nil

		case "down", "j":
			m.cursor = (m.cursor + 1) % len(menuItems)
			m.status = ""
			return m, nil

		case "enter":
			return m.selectItem()
		}
	}

	return m, nil
}

func (m WelcomeModel) selectItem() (tea.Model, tea.Cmd) {
	item := menuItems[m.cursor]

	if !item.enabled {
		m.status = "“" + item.label + "” ainda não implementado"
		return m, nil
	}

	switch item.label {
	case "Containers":
		next := NewContainersModel()
		return next, next.Init()
	case "Sair":
		return m, tea.Quit
	}

	return m, nil
}

func (m WelcomeModel) View() string {
	width := m.width
	if width <= 0 {
		width = 80
	}

	primary := lipgloss.NewStyle().Foreground(lipgloss.Color("#4D7CFF")).Bold(true)
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA9FF"))
	faint := lipgloss.NewStyle().Faint(true)
	ruleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1F3A8A"))

	art := primary.Render(asciiPaterna)
	rule := ruleStyle.Render(strings.Repeat("─", width))

	credits := fmt.Sprintf("by %s and %s",
		accent.Render("@hugaojanuario"),
		accent.Render("@ViitoJooj"),
	)

	tagline := faint.Render(fmt.Sprintf(
		"container orchestration & observability · built in Go · %s",
		displayVersion(),
	))

	menu := renderMenu(m.cursor, primary, accent, faint)

	footerLeft := accent.Render("github.com/hugaojanuario/paterna")
	footerRight := faint.Render("MIT License")
	footer := alignRow(footerLeft, footerRight, width)

	hint := faint.Render("  ↑↓: navegar  ·  enter: selecionar  ·  q: sair")

	status := ""
	if m.status != "" {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render("  " + m.status)
	}

	parts := []string{
		"",
		art,
		"",
		rule,
		"",
		credits,
		tagline,
		"",
		rule,
		"",
		menu,
		"",
		status,
		rule,
		footer,
		"",
		hint,
	}

	return strings.Join(parts, "\n")
}

func renderMenu(cursor int, primary, accent, faint lipgloss.Style) string {
	lines := make([]string, 0, len(menuItems))

	for i, item := range menuItems {
		marker := "  "
		labelStyle := lipgloss.NewStyle()
		cmdStyle := faint

		if i == cursor {
			marker = primary.Render("▸ ")
			labelStyle = labelStyle.Foreground(lipgloss.Color("#7AA9FF")).Bold(true)
			cmdStyle = accent
		}

		if !item.enabled {
			labelStyle = labelStyle.Faint(true)
		}

		label := labelStyle.Render(padRight(item.label, 16))
		cmd := cmdStyle.Render("$ " + item.command)

		lines = append(lines, "  "+marker+label+"  "+cmd)
	}

	return strings.Join(lines, "\n")
}

func padRight(s string, width int) string {
	if lipgloss.Width(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-lipgloss.Width(s))
}

func alignRow(left, right string, width int) string {
	gap := max(width-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

func displayVersion() string {
	v := version.Version
	if v == "" || v == "dev" {
		return "v0.1.0-alpha"
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}
