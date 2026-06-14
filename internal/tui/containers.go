package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hugaojanuario/Paterna/internal/container"
)

type containerRow struct {
	Info       container.ContainerInfo
	CPU        float64
	MemMB      float64
	MemLimitMB float64
	HasMetrics bool
}

type containersLoadedMsg struct {
	rows []containerRow
}

type errMsg struct {
	err error
}

type actionDoneMsg struct {
	text string
}

type tickMsg time.Time

type ContainersModel struct {
	rows   []containerRow
	cursor int
	width  int
	height int
	status string
	err    error
}

func NewContainersModel() ContainersModel {
	return ContainersModel{}
}

func (m ContainersModel) Init() tea.Cmd {
	return tea.Batch(loadContainers, tickEvery())
}

func (m ContainersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case containersLoadedMsg:
		m.rows = msg.rows
		m.err = nil
		if m.cursor >= len(m.rows) {
			m.cursor = max(len(m.rows)-1, 0)
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case actionDoneMsg:
		m.status = msg.text
		return m, loadContainers

	case tickMsg:
		return m, tea.Batch(loadContainers, tickEvery())

	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			return NewWelcomeModel(), nil

		case "up", "k":
			if len(m.rows) == 0 {
				return m, nil
			}
			m.cursor = (m.cursor - 1 + len(m.rows)) % len(m.rows)
			return m, nil

		case "down", "j":
			if len(m.rows) == 0 {
				return m, nil
			}
			m.cursor = (m.cursor + 1) % len(m.rows)
			return m, nil

		case "enter":
			id := m.selectedID()
			if id == "" {
				return m, nil
			}
			next := NewContainerDetailsModel(id, m.selectedName())
			return next, next.Init()

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

	return m, nil
}

func (m ContainersModel) View() string {
	width := m.width
	if width <= 0 {
		width = 120
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5BB0")).
		Bold(true).
		Render("Paterna — Containers")

	upCount := 0
	for _, r := range m.rows {
		if strings.HasPrefix(strings.ToLower(r.Info.Status), "up") {
			upCount++
		}
	}

	live := lipgloss.NewStyle().Foreground(lipgloss.Color("#00C875")).Render("●") +
		" " + lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d up", upCount)) +
		lipgloss.NewStyle().Faint(true).Render(" · ↻ live (1s)")

	header := alignRow(title, live, width)

	cols := columnWidths(width)
	tableHeader := renderHeaderRow(cols)
	body := renderBody(m.rows, m.cursor, cols)

	hint := lipgloss.NewStyle().Faint(true).Render(
		"↑↓ navegar  enter detalhes  s start  x stop  r restart  u atualizar  esc voltar  q sair",
	)

	status := ""
	if m.err != nil {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5757")).
			Render("erro: " + m.err.Error())
	} else if m.status != "" {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("#00C875")).
			Render(m.status)
	}

	parts := []string{
		"",
		header,
		"",
		tableHeader,
		body,
		"",
		status,
		hint,
		"",
	}

	return strings.Join(parts, "\n")
}

type colWidths struct {
	id, name, image, cpu, mem, status int
}

func columnWidths(termWidth int) colWidths {
	// larguras fixas, soma ≈ 110; sobra vira padding no STATUS
	c := colWidths{id: 14, name: 18, image: 16, cpu: 24, mem: 20, status: 16}
	if termWidth > 120 {
		c.status = termWidth - (c.id + c.name + c.image + c.cpu + c.mem + 10)
	}
	return c
}

func renderHeaderRow(c colWidths) string {
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8B8FA8"))
	return "  " + style.Render(
		padRight("ID", c.id)+
			padRight("NAME", c.name)+
			padRight("IMAGE", c.image)+
			padRight("CPU", c.cpu)+
			padRight("MEM", c.mem)+
			padRight("STATUS", c.status),
	)
}

func renderBody(rows []containerRow, cursor int, c colWidths) string {
	if len(rows) == 0 {
		return "  " + lipgloss.NewStyle().Faint(true).Render("(nenhum container)")
	}

	lines := make([]string, 0, len(rows))
	for i, r := range rows {
		line := renderRow(r, c, i == cursor)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func renderRow(r containerRow, c colWidths, selected bool) string {
	cpuCell := metricCell(r.CPU, fmt.Sprintf("%.1f%%", r.CPU), r.HasMetrics, c.cpu)
	memPct := 0.0
	if r.MemLimitMB > 0 {
		memPct = (r.MemMB / r.MemLimitMB) * 100.0
	}
	memCell := metricCell(memPct, fmt.Sprintf("%.0f MB", r.MemMB), r.HasMetrics, c.mem)

	id := truncate(shortID(r.Info.ID), c.id-1)
	name := truncate(r.Info.Name, c.name-1)
	image := truncate(r.Info.Image, c.image-1)
	status := truncate(r.Info.Status, c.status-1)

	row := "  " +
		padRight(id, c.id) +
		padRight(name, c.name) +
		padRight(image, c.image) +
		padCell(cpuCell, c.cpu) +
		padCell(memCell, c.mem) +
		padRight(status, c.status)

	if selected {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5B5BFF")).
			Bold(true).
			Render(row)
	}
	return row
}

func metricCell(percent float64, label string, hasMetrics bool, width int) string {
	if !hasMetrics {
		return lipgloss.NewStyle().Faint(true).Render("—")
	}

	barWidth := 8
	labelWidth := 7

	color := "#00C875"
	switch {
	case percent >= 75:
		color = "#FF5757"
	case percent >= 40:
		color = "#FFB020"
	}

	filled := int(percent / 100.0 * float64(barWidth))
	filled = clamp(filled, 0, barWidth)

	full := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(strings.Repeat("█", filled))
	empty := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2A2A35")).
		Render(strings.Repeat("░", barWidth-filled))

	bar := full + empty
	value := lipgloss.NewStyle().Bold(true).Render(padRight(label, labelWidth))

	_ = width
	return bar + " " + value
}

func padCell(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func truncate(s string, max int) string {
	if max <= 1 {
		return s
	}
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func tickEvery() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m ContainersModel) selectedID() string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return ""
	}
	return m.rows[m.cursor].Info.ID
}

func (m ContainersModel) selectedName() string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return ""
	}
	return m.rows[m.cursor].Info.Name
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

	enriched := make([]containerRow, len(rows))
	var wg sync.WaitGroup

	for i, c := range rows {
		enriched[i] = containerRow{Info: c}

		if !strings.HasPrefix(strings.ToLower(c.Status), "up") {
			continue
		}

		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			stats, err := container.GetContainerStats(id)
			if err != nil {
				return
			}
			cpu, mem, memLimit := container.ComputeUsage(stats)
			enriched[i].CPU = cpu
			enriched[i].MemMB = mem
			enriched[i].MemLimitMB = memLimit
			enriched[i].HasMetrics = true
		}(i, c.ID)
	}
	wg.Wait()

	return containersLoadedMsg{rows: enriched}
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
