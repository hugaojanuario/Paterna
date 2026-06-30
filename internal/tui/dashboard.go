package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hugaojanuario/Paterna/internal/container"
	"github.com/hugaojanuario/Paterna/internal/system"
)

const histLen = 60

// cores do tema
const (
	colBorder = "#3A3A4A"
	colTitle  = "#7AA9FF"
	colDim    = "#8B8FA8"
	colGreen  = "#00C875"
	colYellow = "#FFB020"
	colRed    = "#FF5757"
	colCyan   = "#36D7E0"
	colPink   = "#FF5BB0"
)

type dashboardDataMsg struct {
	snap       system.Snapshot
	containers []container.ContainerInfo
	dockerErr  error
}

type DashboardModel struct {
	col        *system.Collector
	snap       system.Snapshot
	containers []container.ContainerInfo
	dockerErr  error

	cpuHist []float64
	memHist []float64

	width  int
	height int
}

func NewDashboardModel() DashboardModel {
	return DashboardModel{col: system.NewCollector()}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(collectAll(m.col), tickEvery())
}

func collectAll(col *system.Collector) tea.Cmd {
	return func() tea.Msg {
		snap := col.Collect()
		rows, err := container.List(true)
		return dashboardDataMsg{snap: snap, containers: rows, dockerErr: err}
	}
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case dashboardDataMsg:
		m.snap = msg.snap
		m.containers = msg.containers
		m.dockerErr = msg.dockerErr
		m.cpuHist = pushHist(m.cpuHist, msg.snap.CPUTotal)
		m.memHist = pushHist(m.memHist, msg.snap.MemPercent)
		return m, nil

	case tickMsg:
		return m, tea.Batch(collectAll(m.col), tickEvery())

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "d", "enter":
			next := NewContainersModel()
			return next, next.Init()
		}
	}

	return m, nil
}

func pushHist(h []float64, v float64) []float64 {
	h = append(h, v)
	if len(h) > histLen {
		h = h[len(h)-histLen:]
	}
	return h
}

func (m DashboardModel) View() string {
	width := m.width
	if width <= 0 {
		width = 120
	}

	leftW := width * 6 / 10
	rightW := width - leftW

	cpuBox := m.renderCPU(leftW)
	memBox := m.renderMem(rightW)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, cpuBox, memBox)

	diskBox := m.renderDisks(leftW)
	netBox := m.renderNet(rightW)
	midRow := lipgloss.JoinHorizontal(lipgloss.Top, diskBox, netBox)

	procBox := m.renderProcs(width)
	dockerBox := m.renderDocker(width)

	parts := []string{
		m.renderHeader(width),
		topRow,
		midRow,
		procBox,
		dockerBox,
		m.renderHint(),
	}
	return strings.Join(parts, "\n")
}

func (m DashboardModel) renderHeader(width int) string {
	s := m.snap
	bold := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colPink))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(colDim))

	left := bold.Render("Paterna") + dim.Render(" · "+orDash(s.Hostname)+" · "+orDash(s.OS))

	right := dim.Render(fmt.Sprintf("up %s · load %.2f %.2f %.2f · %s",
		fmtUptime(s.Uptime), s.Load1, s.Load5, s.Load15, displayVersion()))

	return alignRow(left, right, width)
}

func (m DashboardModel) renderCPU(width int) string {
	s := m.snap
	inner := innerWidth(width)

	var b strings.Builder
	b.WriteString(labeledBar("total", s.CPUTotal, inner, pctColor(s.CPUTotal)))
	b.WriteString("\n")
	b.WriteString(sparkline(m.cpuHist, inner, pctColor(s.CPUTotal)))
	b.WriteString("\n\n")

	// cores em duas colunas
	half := (inner - 2) / 2
	cores := s.CPUPerCore
	rows := (len(cores) + 1) / 2
	for r := 0; r < rows; r++ {
		line := coreCell(r, cores, half)
		if r+rows < len(cores) {
			line += "  " + coreCell(r+rows, cores, half)
		}
		b.WriteString(line + "\n")
	}

	return box("CPU", strings.TrimRight(b.String(), "\n"), width, colTitle)
}

func coreCell(i int, cores []float64, width int) string {
	if i >= len(cores) {
		return strings.Repeat(" ", width)
	}
	label := fmt.Sprintf("c%-2d", i)
	barW := width - len(label) - 6
	if barW < 3 {
		barW = 3
	}
	bar := miniBar(cores[i], barW, pctColor(cores[i]))
	val := fmt.Sprintf("%4.0f%%", cores[i])
	return label + bar + val
}

func (m DashboardModel) renderMem(width int) string {
	s := m.snap
	inner := innerWidth(width)

	var b strings.Builder
	b.WriteString(labeledBar("ram", s.MemPercent, inner, pctColor(s.MemPercent)))
	b.WriteString("\n")
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(colDim))
	b.WriteString(dim.Render(fmt.Sprintf("%s / %s", fmtBytes(s.MemUsed), fmtBytes(s.MemTotal))))
	b.WriteString("\n")
	b.WriteString(sparkline(m.memHist, inner, pctColor(s.MemPercent)))
	b.WriteString("\n\n")

	swapPct := 0.0
	if s.SwapTotal > 0 {
		swapPct = float64(s.SwapUsed) / float64(s.SwapTotal) * 100.0
	}
	b.WriteString(labeledBar("swp", swapPct, inner, colCyan))
	b.WriteString("\n")
	b.WriteString(dim.Render(fmt.Sprintf("%s / %s", fmtBytes(s.SwapUsed), fmtBytes(s.SwapTotal))))

	return box("MEM", b.String(), width, colTitle)
}

func (m DashboardModel) renderDisks(width int) string {
	inner := innerWidth(width)
	var b strings.Builder

	if len(m.snap.Disks) == 0 {
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("(sem dados de disco)"))
	}
	for _, d := range m.snap.Disks {
		label := truncate(d.Path, 10)
		b.WriteString(labeledBar(label, d.UsedPercent, inner, pctColor(d.UsedPercent)))
		b.WriteString("\n")
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color(colDim))
		b.WriteString(dim.Render(fmt.Sprintf("  %s / %s", fmtBytes(d.Used), fmtBytes(d.Total))))
		b.WriteString("\n")
	}

	return box("DISK", strings.TrimRight(b.String(), "\n"), width, colTitle)
}

func (m DashboardModel) renderNet(width int) string {
	s := m.snap
	down := lipgloss.NewStyle().Foreground(lipgloss.Color(colGreen)).Bold(true)
	up := lipgloss.NewStyle().Foreground(lipgloss.Color(colYellow)).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(colDim))

	body := strings.Join([]string{
		down.Render("↓ ") + dim.Render("download  ") + down.Render(fmtRate(s.NetRecvRate)),
		up.Render("↑ ") + dim.Render("upload    ") + up.Render(fmtRate(s.NetSentRate)),
	}, "\n\n")

	return box("NET", body, width, colTitle)
}

func (m DashboardModel) renderProcs(width int) string {
	inner := innerWidth(width)
	dim := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colDim))

	nameW := inner - 7 - 8 - 12
	if nameW < 8 {
		nameW = 8
	}

	header := dim.Render(
		padRight("PID", 7) + padRight("NAME", nameW) + padRight("CPU%", 8) + "MEM",
	)

	limit := m.procRows()
	lines := []string{header}
	for i, p := range m.snap.Procs {
		if i >= limit {
			break
		}
		row := padRight(fmt.Sprintf("%d", p.PID), 7) +
			padRight(truncate(p.Name, nameW-1), nameW) +
			padRight(fmt.Sprintf("%.1f", p.CPU), 8) +
			fmt.Sprintf("%.0f MB", p.MemMB)
		lines = append(lines, row)
	}

	return box("PROCESSOS", strings.Join(lines, "\n"), width, colTitle)
}

// procRows decide quantas linhas de processo cabem dado o resto do layout.
func (m DashboardModel) procRows() int {
	rows := m.height - 30
	if rows < 4 {
		rows = 4
	}
	if rows > system.MaxProcs {
		rows = system.MaxProcs
	}
	return rows
}

func (m DashboardModel) renderDocker(width int) string {
	if m.dockerErr != nil {
		body := lipgloss.NewStyle().Foreground(lipgloss.Color(colRed)).
			Render("docker indisponível: " + m.dockerErr.Error())
		return box("DOCKER", body, width, colPink)
	}

	up := 0
	for _, c := range m.containers {
		if strings.HasPrefix(strings.ToLower(c.Status), "up") {
			up++
		}
	}

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(colDim))
	summary := fmt.Sprintf("%s%d up%s · %d total   %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color(colGreen)).Render("● "),
		up,
		"",
		len(m.containers),
		dim.Render("[d/enter] gerenciar containers"),
	)

	lines := []string{summary}
	for i, c := range m.containers {
		if i >= 4 {
			lines = append(lines, dim.Render(fmt.Sprintf("  … +%d", len(m.containers)-4)))
			break
		}
		dot := lipgloss.NewStyle().Foreground(lipgloss.Color(colRed)).Render("●")
		if strings.HasPrefix(strings.ToLower(c.Status), "up") {
			dot = lipgloss.NewStyle().Foreground(lipgloss.Color(colGreen)).Render("●")
		}
		lines = append(lines, fmt.Sprintf("  %s %s %s",
			dot, padRight(truncate(c.Name, 24), 24), dim.Render(truncate(c.Status, 30))))
	}

	return box("DOCKER", strings.Join(lines, "\n"), width, colPink)
}

func (m DashboardModel) renderHint() string {
	return lipgloss.NewStyle().Faint(true).Render(
		"  d/enter: containers  ·  q: sair",
	)
}

// --- helpers de render ---

func box(title, body string, width int, titleColor string) string {
	inner := innerWidth(width)
	t := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(titleColor)).Render(" " + title + " ")

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colBorder)).
		Width(inner).
		Padding(0, 1)

	return style.Render(t + "\n" + body)
}

func innerWidth(width int) int {
	w := width - 4 // borda (2) + padding (2)
	if w < 10 {
		w = 10
	}
	return w
}

// labeledBar: "label [████░░░░] 42%"
func labeledBar(label string, pct float64, width int, color string) string {
	lbl := padRight(label, 6)
	tail := 6 // espaço pra " 100%"
	barW := width - len(lbl) - tail
	if barW < 4 {
		barW = 4
	}
	return lbl + miniBar(pct, barW, color) + fmt.Sprintf("%4.0f%%", pct)
}

func miniBar(pct float64, width int, color string) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct / 100.0 * float64(width))
	filled = clamp(filled, 0, width)

	full := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).
		Render(strings.Repeat("█", filled))
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A2A35")).
		Render(strings.Repeat("░", width-filled))
	return "[" + full + empty + "] "
}

var sparkRunes = []rune("▁▂▃▄▅▆▇█")

func sparkline(hist []float64, width int, color string) string {
	if width < 1 {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	data := hist
	if len(data) > width {
		data = data[len(data)-width:]
	}

	var b strings.Builder
	pad := width - len(data)
	for i := 0; i < pad; i++ {
		b.WriteRune(' ')
	}
	for _, v := range data {
		idx := int(v / 100.0 * float64(len(sparkRunes)-1))
		idx = clamp(idx, 0, len(sparkRunes)-1)
		b.WriteRune(sparkRunes[idx])
	}
	return style.Render(b.String())
}

func pctColor(pct float64) string {
	switch {
	case pct >= 80:
		return colRed
	case pct >= 50:
		return colYellow
	default:
		return colGreen
	}
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func fmtBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func fmtRate(bytesPerSec float64) string {
	if bytesPerSec < 0 {
		bytesPerSec = 0
	}
	return fmtBytes(uint64(bytesPerSec)) + "/s"
}

func fmtUptime(d time.Duration) string {
	if d <= 0 {
		return "—"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
