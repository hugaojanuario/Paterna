package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/hugaojanuario/Paterna/internal/container"
)

type detailsLoadedMsg struct {
	state container.ContainerState
	cpu   float64
	mem   float64
}

type logStreamReadyMsg struct {
	ch chan string
}

type logLineMsg struct {
	line string
	ok   bool
}

type logStreamErrMsg struct {
	err error
}

type ContainerDetailsModel struct {
	id     string
	name   string
	state  container.ContainerState
	cpu    float64
	mem    float64
	hasMet bool
	logs   []string
	logCh  chan string
	cancel context.CancelFunc

	viewport viewport.Model
	width    int
	height   int
	status   string
	err      error
}

func NewContainerDetailsModel(id, name string) ContainerDetailsModel {
	vp := viewport.New(0, 0)
	vp.SetContent("(carregando logs...)")
	return ContainerDetailsModel{
		id:       id,
		name:     name,
		viewport: vp,
	}
}

func (m ContainerDetailsModel) Init() tea.Cmd {
	return tea.Batch(
		loadDetails(m.id),
		startLogStream(m.id),
		tickEvery(),
	)
}

func (m ContainerDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = max(msg.Height-16, 6)
		return m, nil

	case detailsLoadedMsg:
		m.state = msg.state
		m.cpu = msg.cpu
		m.mem = msg.mem
		m.hasMet = msg.cpu > 0 || msg.mem > 0
		return m, nil

	case tickMsg:
		return m, tea.Batch(loadDetails(m.id), tickEvery())

	case logStreamReadyMsg:
		m.logCh = msg.ch
		return m, waitForLogLine(msg.ch)

	case logLineMsg:
		if !msg.ok {
			return m, nil
		}
		m.logs = append(m.logs, msg.line)
		if len(m.logs) > 500 {
			m.logs = m.logs[len(m.logs)-500:]
		}
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		return m, waitForLogLine(m.logCh)

	case logStreamErrMsg:
		m.err = msg.err
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case actionDoneMsg:
		m.status = msg.text
		return m, loadDetails(m.id)

	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			m.stopStream()
			return m, tea.Quit

		case "esc":
			m.stopStream()
			next := NewContainersModel()
			return next, next.Init()

		case "s":
			m.status = "iniciando..."
			return m, startContainer(m.id)

		case "x":
			m.status = "parando..."
			return m, stopContainer(m.id)

		case "r":
			m.status = "reiniciando..."
			return m, restartContainer(m.id)
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *ContainerDetailsModel) stopStream() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}

func (m ContainerDetailsModel) View() string {
	width := m.width
	if width <= 0 {
		width = 120
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5BB0")).
		Bold(true).
		Render("Paterna — " + m.name)

	statusColor := "#00C875"
	if !strings.HasPrefix(strings.ToLower(m.state.Status), "running") &&
		!strings.HasPrefix(strings.ToLower(m.state.Status), "up") {
		statusColor = "#FF5757"
	}
	statusBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Bold(true).
		Render("● " + safeStatus(m.state.Status))

	header := alignRow(title, statusBadge, width)

	info := renderDetailsInfo(m)
	logsHeader := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8B8FA8")).
		Bold(true).
		Render("LOGS")

	hint := lipgloss.NewStyle().Faint(true).Render(
		"↑↓ scroll  s start  x stop  r restart  esc voltar  q sair",
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
		info,
		"",
		logsHeader,
		m.viewport.View(),
		"",
		status,
		hint,
		"",
	}

	return strings.Join(parts, "\n")
}

func renderDetailsInfo(m ContainerDetailsModel) string {
	faint := lipgloss.NewStyle().Faint(true)
	bold := lipgloss.NewStyle().Bold(true)

	uptime := "—"
	if !m.state.FinishedAt.IsZero() {
		uptime = m.state.FinishedAt.Format(time.RFC3339)
	}

	cpuStr := "—"
	memStr := "—"
	if m.hasMet {
		cpuStr = fmt.Sprintf("%.1f%%", m.cpu)
		memStr = fmt.Sprintf("%.0f MB", m.mem)
	}

	lines := []string{
		faint.Render("  ID:        ") + bold.Render(shortID(m.id)),
		faint.Render("  Nome:      ") + bold.Render(m.name),
		faint.Render("  Status:    ") + bold.Render(safeStatus(m.state.Status)),
		faint.Render("  CPU:       ") + bold.Render(cpuStr) +
			faint.Render("   MEM: ") + bold.Render(memStr),
		faint.Render("  Encerrado: ") + bold.Render(uptime),
	}

	if m.state.OOMKilled {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5757")).
			Render("  ⚠ OOM killed"))
	}

	return strings.Join(lines, "\n")
}

func safeStatus(s string) string {
	if s == "" {
		return "(desconhecido)"
	}
	return s
}

func loadDetails(id string) tea.Cmd {
	return func() tea.Msg {
		state, err := container.InspectContainer(id)
		if err != nil {
			return errMsg{err: err}
		}

		var cpu, mem float64
		if strings.EqualFold(state.Status, "running") {
			if stats, err := container.GetContainerStats(id); err == nil {
				cpu, mem, _ = container.ComputeUsage(stats)
			}
		}

		return detailsLoadedMsg{state: state, cpu: cpu, mem: mem}
	}
}

func startLogStream(id string) tea.Cmd {
	return func() tea.Msg {
		rc, err := container.StreamContainerLogs(id)
		if err != nil {
			return logStreamErrMsg{err: err}
		}

		ctx, cancel := context.WithCancel(context.Background())
		ch := make(chan string, 64)

		go pumpLogs(ctx, rc, ch)

		_ = cancel // cancel ainda não está conectado ao model

		return logStreamReadyMsg{ch: ch}
	}
}

// pumpLogs demuxa logs do Docker e empurra linhas no channel até EOF/cancel
func pumpLogs(ctx context.Context, rc io.ReadCloser, ch chan<- string) {
	defer close(ch)
	defer rc.Close()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		_, _ = stdcopy.StdCopy(pw, pw, rc)
	}()

	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case ch <- scanner.Text():
		}
	}
}

func waitForLogLine(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		return logLineMsg{line: line, ok: ok}
	}
}
