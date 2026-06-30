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

const histLen = 120

// paleta do tema (dark, estilo btop)
const (
	colBorder = "#2C2C3A"
	colDim    = "#6C7086"
	colText   = "#CDD6F4"
	colGreen  = "#00C875"
	colYellow = "#FFB020"
	colRed    = "#FF5757"
	colCyan   = "#36D7E0"
	colBlue   = "#5B8DFF"
	colPurple = "#B47AFF"
	colPink   = "#FF5BB0"
	colEmpty  = "#23232F"
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

	cpuHist  []float64
	memHist  []float64
	recvHist []float64
	sentHist []float64

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
		m.recvHist = pushHist(m.recvHist, msg.snap.NetRecvRate)
		m.sentHist = pushHist(m.sentHist, msg.snap.NetSentRate)
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

// ============================ View ============================

func (m DashboardModel) View() string {
	width := m.width
	if width <= 0 {
		width = 120
	}

	leftW := width * 62 / 100
	rightW := width - leftW

	var out []string
	out = append(out, m.renderTopline(width))

	// linha 1: CPU | MEM
	cpu := boxSpec{1, "cpu", colCyan, m.cpuLines(innerWidth(leftW))}
	mem := boxSpec{2, "mem", colGreen, m.memLines(innerWidth(rightW))}
	out = append(out, renderBoxRow(cpu, mem, leftW, rightW))

	// linha 2: DISK | NET
	disk := boxSpec{3, "disk", colYellow, m.diskLines(innerWidth(leftW))}
	net := boxSpec{4, "net", colBlue, m.netLines(innerWidth(rightW))}
	out = append(out, renderBoxRow(disk, net, leftW, rightW))

	// linha 3: PROCESSOS (full)
	proc := boxSpec{5, "proc", colPurple, m.procLines(innerWidth(width))}
	out = append(out, drawBox(proc, width, len(proc.lines)))

	// linha 4: DOCKER (full)
	dock := boxSpec{6, "docker", colPink, m.dockerLines(innerWidth(width))}
	out = append(out, drawBox(dock, width, len(dock.lines)))

	out = append(out, m.renderHint())
	return strings.Join(out, "\n")
}

func (m DashboardModel) renderTopline(width int) string {
	s := m.snap
	name := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colPink)).Render("paterna")
	host := lipgloss.NewStyle().Foreground(lipgloss.Color(colText)).Render(orDash(s.Hostname))
	os := lipgloss.NewStyle().Foreground(lipgloss.Color(colDim)).Render(orDash(s.OS))
	left := name + " " + host + " " + dim("· "+os)

	clock := time.Now().Format("15:04:05")
	right := dim("up "+fmtUptime(s.Uptime)+" · ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(colText)).Render(clock) +
		dim(" · "+displayVersion())

	return " " + alignRow(left, right, width-2)
}

func (m DashboardModel) renderHint() string {
	return dim("  d/enter containers · q sair")
}

// ============================ CPU ============================

func (m DashboardModel) cpuLines(inner int) []string {
	s := m.snap
	var lines []string

	// medidor total + gráfico de área
	lines = append(lines, meterRow("CPU", s.CPUTotal, inner))
	graph := areaGraph(m.cpuHist, inner, 4)
	lines = append(lines, graph...)
	lines = append(lines, "")

	// grade de cores em colunas
	lines = append(lines, coreGrid(s.CPUPerCore, inner)...)

	// load average
	lines = append(lines, "")
	lines = append(lines, dim("load avg ")+
		valStyle(fmt.Sprintf("%.2f", s.Load1))+dim("  ")+
		valStyle(fmt.Sprintf("%.2f", s.Load5))+dim("  ")+
		valStyle(fmt.Sprintf("%.2f", s.Load15)))

	return lines
}

func coreGrid(cores []float64, inner int) []string {
	if len(cores) == 0 {
		return []string{dim("(sem dados de cpu)")}
	}

	const cell = 24
	cols := inner / cell
	if cols < 1 {
		cols = 1
	}
	if cols > 4 {
		cols = 4
	}
	rows := (len(cores) + cols - 1) / cols

	out := make([]string, 0, rows)
	for r := 0; r < rows; r++ {
		var parts []string
		for c := 0; c < cols; c++ {
			i := c*rows + r
			if i >= len(cores) {
				parts = append(parts, strings.Repeat(" ", cell-2))
				continue
			}
			parts = append(parts, coreCell(i, cores[i], cell-2))
		}
		out = append(out, strings.Join(parts, "  "))
	}
	return out
}

func coreCell(i int, pct float64, width int) string {
	label := dim(fmt.Sprintf("c%-2d", i))
	val := pctStyle(pct).Render(fmt.Sprintf("%4.0f%%", pct))
	barW := width - 3 - 6
	if barW < 3 {
		barW = 3
	}
	return label + gradMeter(pct, barW) + " " + val
}

// ============================ MEM ============================

func (m DashboardModel) memLines(inner int) []string {
	s := m.snap
	var lines []string

	lines = append(lines, meterRow("RAM", s.MemPercent, inner))
	lines = append(lines, dim("  "+fmtBytes(s.MemUsed))+dim(" / ")+valStyle(fmtBytes(s.MemTotal)))
	lines = append(lines, areaGraph(m.memHist, inner, 3)...)
	lines = append(lines, "")

	swapPct := 0.0
	if s.SwapTotal > 0 {
		swapPct = float64(s.SwapUsed) / float64(s.SwapTotal) * 100.0
	}
	lines = append(lines, meterRow("SWP", swapPct, inner))
	lines = append(lines, dim("  "+fmtBytes(s.SwapUsed))+dim(" / ")+valStyle(fmtBytes(s.SwapTotal)))

	return lines
}

// ============================ DISK ============================

func (m DashboardModel) diskLines(inner int) []string {
	if len(m.snap.Disks) == 0 {
		return []string{dim("(sem dados de disco)")}
	}
	var lines []string
	for _, d := range m.snap.Disks {
		head := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colText)).
			Render(truncate(d.Path, 14))
		used := valStyle(fmtBytes(d.Used) + " / " + fmtBytes(d.Total))
		lines = append(lines, alignVis(head, used, inner))
		lines = append(lines, meterRow("", d.UsedPercent, inner))
	}
	return lines
}

// ============================ NET ============================

func (m DashboardModel) netLines(inner int) []string {
	s := m.snap
	down := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colGreen))
	up := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colYellow))

	var lines []string
	lines = append(lines, alignVis(down.Render("▼ download"), down.Render(fmtRate(s.NetRecvRate)), inner))
	lines = append(lines, autoGraph(m.recvHist, inner, 2, colGreen)...)
	lines = append(lines, "")
	lines = append(lines, alignVis(up.Render("▲ upload"), up.Render(fmtRate(s.NetSentRate)), inner))
	lines = append(lines, autoGraph(m.sentHist, inner, 2, colYellow)...)
	return lines
}

// ============================ PROC ============================

func (m DashboardModel) procLines(inner int) []string {
	head := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colDim))
	nameW := inner - 8 - 9 - 11
	if nameW < 8 {
		nameW = 8
	}

	header := head.Render(padRight("PID", 8) + padRight("NAME", nameW) +
		padRight("CPU%", 9) + "MEM")
	lines := []string{header}

	limit := m.procRows()
	for i, p := range m.snap.Procs {
		if i >= limit {
			break
		}
		row := dim(padRight(fmt.Sprintf("%d", p.PID), 8)) +
			lipgloss.NewStyle().Foreground(lipgloss.Color(colText)).Render(padRight(truncate(p.Name, nameW-1), nameW)) +
			pctStyle(p.CPU).Render(padRight(fmt.Sprintf("%.1f", p.CPU), 9)) +
			valStyle(fmt.Sprintf("%.0f MB", p.MemMB))
		lines = append(lines, row)
	}
	return lines
}

func (m DashboardModel) procRows() int {
	rows := m.height - 34
	if rows < 4 {
		rows = 4
	}
	if rows > system.MaxProcs {
		rows = system.MaxProcs
	}
	return rows
}

// ============================ DOCKER ============================

func (m DashboardModel) dockerLines(inner int) []string {
	if m.dockerErr != nil {
		return []string{lipgloss.NewStyle().Foreground(lipgloss.Color(colRed)).
			Render("docker indisponível: " + m.dockerErr.Error())}
	}

	up := 0
	for _, c := range m.containers {
		if isUp(c.Status) {
			up++
		}
	}

	summary := lipgloss.NewStyle().Foreground(lipgloss.Color(colGreen)).Render("● ") +
		valStyle(fmt.Sprintf("%d", up)) + dim(" up · ") +
		valStyle(fmt.Sprintf("%d", len(m.containers))) + dim(" total")
	hint := dim("[d/enter] gerenciar")
	lines := []string{alignVis(summary, hint, inner)}

	for i, c := range m.containers {
		if i >= 5 {
			lines = append(lines, dim(fmt.Sprintf("  … +%d", len(m.containers)-5)))
			break
		}
		dotColor := colRed
		if isUp(c.Status) {
			dotColor = colGreen
		}
		dot := lipgloss.NewStyle().Foreground(lipgloss.Color(dotColor)).Render("●")
		name := lipgloss.NewStyle().Foreground(lipgloss.Color(colText)).Render(padRight(truncate(c.Name, 26), 26))
		lines = append(lines, "  "+dot+" "+name+dim(truncate(c.Status, inner-32)))
	}
	return lines
}

// ============================ caixas ============================

type boxSpec struct {
	num    int
	title  string
	accent string
	lines  []string
}

// renderBoxRow desenha duas caixas lado a lado com a MESMA altura, garantindo
// que as bordas fiquem alinhadas.
func renderBoxRow(left, right boxSpec, leftW, rightW int) string {
	h := max(len(left.lines), len(right.lines))
	l := drawBox(left, leftW, h)
	r := drawBox(right, rightW, h)
	return lipgloss.JoinHorizontal(lipgloss.Top, l, r)
}

var superscripts = []rune("⁰¹²³⁴⁵⁶⁷⁸⁹")

// drawBox desenha a caixa com título embutido na borda superior (estilo btop)
// e altura de conteúdo fixa (contentH linhas).
func drawBox(b boxSpec, width, contentH int) string {
	innerW := width - 2
	bs := lipgloss.NewStyle().Foreground(lipgloss.Color(colBorder))

	num := ""
	if b.num >= 0 && b.num < len(superscripts) {
		num = string(superscripts[b.num])
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(b.accent)).
		Render(num + b.title)
	titleW := lipgloss.Width(title)

	dashes := innerW - 1 - titleW
	if dashes < 0 {
		dashes = 0
	}
	top := bs.Render("╭─") + title + bs.Render(strings.Repeat("─", dashes)+"╮")
	bottom := bs.Render("╰" + strings.Repeat("─", innerW) + "╯")

	rows := make([]string, 0, contentH+2)
	rows = append(rows, top)
	for i := 0; i < contentH; i++ {
		body := ""
		if i < len(b.lines) {
			body = b.lines[i]
		}
		rows = append(rows, bs.Render("│")+padVis(" "+body, innerW)+bs.Render("│"))
	}
	rows = append(rows, bottom)
	return strings.Join(rows, "\n")
}

func innerWidth(width int) int {
	w := width - 4 // 2 bordas + 1 espaço de padding interno (esq) + 1 folga (dir)
	if w < 10 {
		w = 10
	}
	return w
}

// ============================ medidores / gráficos ============================

// meterRow: "LABEL ▕████████▏  42%" com gradiente e valor à direita.
func meterRow(label string, pct float64, inner int) string {
	lbl := ""
	lblW := 0
	if label != "" {
		lbl = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colDim)).Render(padRight(label, 4))
		lblW = 4
	}
	valTxt := fmt.Sprintf("%4.0f%%", pct)
	val := pctStyle(pct).Render(valTxt)

	barW := inner - lblW - len(valTxt) - 2
	if barW < 4 {
		barW = 4
	}
	return lbl + gradMeter(pct, barW) + " " + val
}

// gradMeter desenha uma barra preenchida com gradiente verde→amarelo→vermelho
// ao longo da extensão (não só pela cor final), como no btop.
func gradMeter(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct/100.0*float64(width) + 0.5)
	filled = clamp(filled, 0, width)

	var b strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			frac := 0.0
			if width > 1 {
				frac = float64(i) / float64(width-1)
			}
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(gradHex(frac))).Render("█"))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colEmpty)).Render("░"))
		}
	}
	return b.String()
}

var blockLevels = []rune(" ▁▂▃▄▅▆▇█")

// areaGraph desenha um gráfico de área de `height` linhas, escala 0-100,
// cada coluna colorida pelo seu valor (gradiente).
func areaGraph(hist []float64, width, height int) []string {
	return scaledGraph(hist, width, height, 100, "")
}

// autoGraph: igual ao areaGraph mas escala pelo máximo da janela (para taxas
// de rede que não são 0-100) e usa cor fixa.
func autoGraph(hist []float64, width, height int, color string) []string {
	maxV := 1.0
	for _, v := range hist {
		if v > maxV {
			maxV = v
		}
	}
	return scaledGraph(hist, width, height, maxV, color)
}

func scaledGraph(hist []float64, width, height int, scale float64, fixedColor string) []string {
	lines := make([]string, height)
	if width < 1 || height < 1 {
		return lines
	}

	data := hist
	if len(data) > width {
		data = data[len(data)-width:]
	}
	pad := width - len(data)

	for row := 0; row < height; row++ {
		// row 0 é o topo; cada linha cobre uma faixa de 8 "oitavos"
		var b strings.Builder
		for i := 0; i < pad; i++ {
			b.WriteRune(' ')
		}
		for _, v := range data {
			ratio := v / scale
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}
			totalEighths := ratio * float64(height) * 8
			rowFromBottom := height - 1 - row
			cellEighths := totalEighths - float64(rowFromBottom*8)
			level := clamp(int(cellEighths+0.0), 0, 8)
			r := blockLevels[level]

			color := fixedColor
			if color == "" {
				color = gradHex(v / 100.0)
			}
			if r == ' ' {
				b.WriteRune(' ')
			} else {
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(string(r)))
			}
		}
		lines[row] = b.String()
	}
	return lines
}

// ============================ helpers de cor/estilo ============================

func dim(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(colDim)).Render(s)
}

func valStyle(s string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colText)).Render(s)
}

func pctStyle(pct float64) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(pctColor(pct)))
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

// gradHex interpola verde→amarelo→vermelho. frac 0=verde, .5=amarelo, 1=vermelho.
func gradHex(frac float64) string {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	var r, g, bl int
	if frac < 0.5 {
		t := frac / 0.5
		r = lerp(0x00, 0xFF, t)
		g = lerp(0xC8, 0xB0, t)
		bl = lerp(0x64, 0x20, t)
	} else {
		t := (frac - 0.5) / 0.5
		r = lerp(0xFF, 0xFF, t)
		g = lerp(0xB0, 0x57, t)
		bl = lerp(0x20, 0x57, t)
	}
	return fmt.Sprintf("#%02X%02X%02X", r, g, bl)
}

func lerp(a, b int, t float64) int {
	return a + int(float64(b-a)*t+0.5)
}

// ============================ helpers de largura ============================

// padVis preenche/corta uma string (que pode ter ANSI) até n colunas visuais.
func padVis(s string, n int) string {
	w := lipgloss.Width(s)
	if w == n {
		return s
	}
	if w < n {
		return s + strings.Repeat(" ", n-w)
	}
	return s // não corta conteúdo estilizado; deixa transbordar (raro)
}

// alignVis alinha left à esquerda e right à direita dentro de width colunas visuais.
func alignVis(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// ============================ formatação ============================

func isUp(status string) bool {
	return strings.HasPrefix(strings.ToLower(status), "up")
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
