package app

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/QurieGLord/QTop/internal/providers"
	"github.com/QurieGLord/QTop/internal/ui"
)

const tickInterval = time.Second
const animationInterval = 80 * time.Millisecond
const historyLimit = 60
const layoutGap = 1
const footerHeight = 2
const animationStep = 0.35
const focusAnimationStep = 0.28

type tickMsg time.Time
type animationMsg time.Time

// fetchMsg contains the fetched statistics.
type fetchMsg struct {
	cpuLoad float64
	memStat providers.MemoryStats
	gpus    []providers.GPUStats
	disks   []providers.DiskStats
	procs   []providers.ProcessInfo
	err     error
}

type focusPanel int

const (
	focusCPU focusPanel = iota
	focusGPU
	focusRAM
	focusDisk
	focusProcess
	focusCount // used for modulo
)

type layoutMode int

const (
	layoutCompact layoutMode = iota
	layoutNarrow
	layoutWide
)

type panelRect struct {
	X int
	Y int
	W int
	H int
}

func (r panelRect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

type layoutSpec struct {
	mode layoutMode
	cpu  panelRect
	gpu  panelRect
	ram  panelRect
	disk panelRect
	proc panelRect
}

func (l layoutSpec) panelAt(x, y int) (focusPanel, bool) {
	switch {
	case l.cpu.Contains(x, y):
		return focusCPU, true
	case l.gpu.Contains(x, y):
		return focusGPU, true
	case l.ram.Contains(x, y):
		return focusRAM, true
	case l.disk.Contains(x, y):
		return focusDisk, true
	case l.proc.Contains(x, y):
		return focusProcess, true
	default:
		return focusCPU, false
	}
}

// Model represents the application state.
type Model struct {
	cpuProvider     providers.CPUProvider
	memProvider     providers.MemoryProvider
	gpuProvider     providers.GPUProvider
	diskProvider    providers.DiskProvider
	processProvider providers.ProcessProvider

	cpuLoad float64
	cpuHist []float64

	cpuDisplayLoad float64
	cpuDisplayHist []float64

	memStat providers.MemoryStats
	memHist []float64

	memDisplayPercent  float64
	swapDisplayPercent float64
	memDisplayHist     []float64

	gpus    []providers.GPUStats
	gpuHist []float64

	gpuDisplayLoad float64
	gpuDisplayHist []float64

	disks []providers.DiskStats

	procs           []providers.ProcessInfo
	procScroll      int
	selectedProcIdx int

	focused focusPanel

	err          error
	windowWidth  int
	windowHeight int
	layout       layoutSpec
	status       string
	focusLevels  [focusCount]float64
}

// NewModel creates a new application model.
func NewModel() *Model {
	return &Model{
		cpuProvider:     providers.NewCPUProvider(),
		memProvider:     providers.NewMemoryProvider(),
		gpuProvider:     providers.NewGPUProvider(),
		diskProvider:    providers.NewDiskProvider(),
		processProvider: providers.NewProcessProvider(),
		focused:         focusCPU,
		focusLevels:     [focusCount]float64{1, 0, 0, 0, 0},
	}
}

// Init initializes the application.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnableMouseCellMotion,
		m.tick(),
		m.animate(),
		m.fetchData(),
	)
}

// Update handles incoming messages and updates the state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			m.focused = (m.focused + 1) % focusCount
		case "shift+tab":
			m.focused--
			if m.focused < 0 {
				m.focused = focusCount - 1
			}
		case "up":
			if m.focused == focusProcess {
				m.moveSelection(-1)
			}
		case "down", "j":
			if m.focused == focusProcess {
				m.moveSelection(1)
			}
		case "k":
			if m.focused == focusProcess {
				m.status = m.signalSelectedProcess(syscall.SIGTERM)
			}
		case "K":
			if m.focused == focusProcess {
				m.status = m.signalSelectedProcess(syscall.SIGKILL)
			}
		}
	case tea.MouseMsg:
		if m.layout.mode == layoutCompact {
			break
		}

		if msg.Type == tea.MouseWheelUp && (m.focused == focusProcess || m.layout.proc.Contains(msg.X, msg.Y)) {
			m.focused = focusProcess
			m.moveSelection(-1)
		} else if msg.Type == tea.MouseWheelDown && (m.focused == focusProcess || m.layout.proc.Contains(msg.X, msg.Y)) {
			m.focused = focusProcess
			m.moveSelection(1)
		} else if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if panel, ok := m.layout.panelAt(msg.X, msg.Y); ok {
				m.focused = panel
				if panel == focusProcess {
					if idx, ok := ui.ProcessIndexFromY(m.layout.proc.H, m.procScroll, msg.Y-m.layout.proc.Y); ok {
						if idx < len(m.procs) {
							m.selectedProcIdx = idx
							m.clampProcessSelection()
						}
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.layout = calculateLayout(msg.Width, msg.Height-footerHeight, len(m.disks))
		m.clampProcessSelection()
	case tickMsg:
		return m, tea.Batch(m.tick(), m.fetchData())
	case animationMsg:
		m.advanceAnimations()
		return m, m.animate()
	case fetchMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.cpuLoad = msg.cpuLoad
			m.cpuHist = appendHist(m.cpuHist, m.cpuLoad)
			if len(m.cpuDisplayHist) == 0 {
				m.cpuDisplayLoad = m.cpuLoad
				m.cpuDisplayHist = append([]float64(nil), m.cpuHist...)
			}

			m.memStat = msg.memStat
			m.memHist = appendHist(m.memHist, m.memStat.UsedPercent)
			if len(m.memDisplayHist) == 0 {
				m.memDisplayPercent = m.memStat.UsedPercent
				m.swapDisplayPercent = m.memStat.SwapUsedPercent
				m.memDisplayHist = append([]float64(nil), m.memHist...)
			}

			m.gpus = msg.gpus
			if len(m.gpus) > 0 {
				m.gpuHist = appendHist(m.gpuHist, m.gpus[0].LoadPercent)
				if len(m.gpuDisplayHist) == 0 {
					m.gpuDisplayLoad = m.gpus[0].LoadPercent
					m.gpuDisplayHist = append([]float64(nil), m.gpuHist...)
				}
			} else {
				m.gpuHist = nil
				m.gpuDisplayHist = nil
				m.gpuDisplayLoad = 0
			}

			m.disks = msg.disks
			m.layout = calculateLayout(m.windowWidth, maxInt(1, m.windowHeight-footerHeight), len(m.disks))
			currentPID := m.selectedPID()
			m.procs = msg.procs
			m.restoreSelection(currentPID)
		}
	}
	return m, nil
}

func appendHist(hist []float64, val float64) []float64 {
	hist = append(hist, val)
	if len(hist) > historyLimit {
		hist = hist[len(hist)-historyLimit:]
	}
	return hist
}

// View renders the UI.
func (m *Model) View() string {
	if m.windowWidth == 0 {
		return "Initializing..."
	}

	if m.err != nil {
		return lipgloss.Place(
			m.windowWidth,
			m.windowHeight,
			lipgloss.Left,
			lipgloss.Top,
			fmt.Sprintf("Error: %v\nPress q to quit.", m.err),
		)
	}

	return m.Layout(m.windowWidth, m.windowHeight)
}

// Layout returns the final responsive application view for the given terminal size.
func (m *Model) Layout(width, height int) string {
	contentHeight := maxInt(1, height-footerHeight)
	m.layout = calculateLayout(width, contentHeight, len(m.disks))
	if m.layout.mode == layoutCompact {
		return lipgloss.JoinVertical(lipgloss.Left, m.compactView(width, contentHeight), m.renderFooter(width))
	}

	cpuView := ui.CPUView{
		ComponentState: ui.ComponentState{Focused: m.focused == focusCPU, FocusLevel: m.focusLevels[focusCPU], Width: m.layout.cpu.W, Height: m.layout.cpu.H},
		LoadPercent:    m.cpuDisplayLoad,
		History:        m.cpuDisplayHist,
	}
	gpuView := ui.GPUView{
		ComponentState: ui.ComponentState{Focused: m.focused == focusGPU, FocusLevel: m.focusLevels[focusGPU], Width: m.layout.gpu.W, Height: m.layout.gpu.H},
		GPUs:           m.gpus,
		History:        m.gpuDisplayHist,
	}
	memView := ui.MemoryView{
		ComponentState:  ui.ComponentState{Focused: m.focused == focusRAM, FocusLevel: m.focusLevels[focusRAM], Width: m.layout.ram.W, Height: m.layout.ram.H},
		Total:           m.memStat.Total,
		Used:            m.memStat.Used,
		UsedPercent:     m.memDisplayPercent,
		SwapTotal:       m.memStat.SwapTotal,
		SwapUsed:        m.memStat.SwapUsed,
		SwapUsedPercent: m.swapDisplayPercent,
		History:         m.memDisplayHist,
	}
	diskView := ui.DiskView{
		ComponentState: ui.ComponentState{Focused: m.focused == focusDisk, FocusLevel: m.focusLevels[focusDisk], Width: m.layout.disk.W, Height: m.layout.disk.H},
		Disks:          m.disks,
	}
	procView := ui.ProcessView{
		ComponentState: ui.ComponentState{Focused: m.focused == focusProcess, FocusLevel: m.focusLevels[focusProcess], Width: m.layout.proc.W, Height: m.layout.proc.H},
		Processes:      m.procs,
		Scroll:         m.procScroll,
		Selected:       m.selectedProcIdx,
	}

	var grid string
	switch m.layout.mode {
	case layoutWide:
		topRow := joinHorizontalGap(layoutGap, cpuView.Render(), gpuView.Render())
		midRow := joinHorizontalGap(layoutGap, memView.Render(), diskView.Render())
		grid = joinVerticalGap(layoutGap, topRow, midRow, procView.Render())
	default:
		grid = joinVerticalGap(layoutGap, cpuView.Render(), gpuView.Render(), memView.Render(), diskView.Render(), procView.Render())
	}

	grid = lipgloss.Place(width, contentHeight, lipgloss.Left, lipgloss.Top, grid)
	return lipgloss.JoinVertical(lipgloss.Left, grid, m.renderFooter(width))
}

// tick returns a command that triggers a tickMsg after the interval.
func (m *Model) tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) animate() tea.Cmd {
	return tea.Tick(animationInterval, func(t time.Time) tea.Msg {
		return animationMsg(t)
	})
}

// fetchData retrieves the stats non-blockingly.
func (m *Model) fetchData() tea.Cmd {
	return func() tea.Msg {
		cpuLoad, _ := m.cpuProvider.GetTotalLoad()
		memStat, _ := m.memProvider.GetStats()
		gpus, _ := m.gpuProvider.GetStats()
		disks, _ := m.diskProvider.GetStats()
		procs, _ := m.processProvider.GetTopProcesses(50)

		return fetchMsg{
			cpuLoad: cpuLoad,
			memStat: memStat,
			gpus:    gpus,
			disks:   disks,
			procs:   procs,
			err:     nil, // In a real app we might handle specific errors, but we don't want to crash on missing GPU
		}
	}
}

func (m *Model) clampProcessSelection() {
	if len(m.procs) == 0 {
		m.selectedProcIdx = 0
		m.procScroll = 0
		return
	}

	if m.selectedProcIdx < 0 {
		m.selectedProcIdx = 0
	}
	if m.selectedProcIdx >= len(m.procs) {
		m.selectedProcIdx = len(m.procs) - 1
	}

	visibleRows := ui.ProcessVisibleRows(m.layout.proc.H)
	if visibleRows <= 0 {
		m.procScroll = 0
		return
	}

	if m.selectedProcIdx < m.procScroll {
		m.procScroll = m.selectedProcIdx
	}
	if m.selectedProcIdx >= m.procScroll+visibleRows {
		m.procScroll = m.selectedProcIdx - visibleRows + 1
	}

	maxScroll := len(m.procs) - visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.procScroll > maxScroll {
		m.procScroll = maxScroll
	}
	if m.procScroll < 0 {
		m.procScroll = 0
	}
}

func (m *Model) moveSelection(delta int) {
	m.selectedProcIdx += delta
	m.clampProcessSelection()
}

func (m *Model) selectedPID() int32 {
	if len(m.procs) == 0 || m.selectedProcIdx < 0 || m.selectedProcIdx >= len(m.procs) {
		return 0
	}
	return m.procs[m.selectedProcIdx].PID
}

func (m *Model) restoreSelection(pid int32) {
	if pid != 0 {
		for i, proc := range m.procs {
			if proc.PID == pid {
				m.selectedProcIdx = i
				m.clampProcessSelection()
				return
			}
		}
	}
	m.clampProcessSelection()
}

func (m *Model) signalSelectedProcess(sig syscall.Signal) string {
	if len(m.procs) == 0 {
		return "No process selected"
	}

	pid := m.selectedPID()
	if pid == 0 {
		return "No process selected"
	}

	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return fmt.Sprintf("Process %d not found", pid)
	}
	if err := proc.Signal(sig); err != nil {
		if err == syscall.EPERM {
			return "Access Denied"
		}
		return fmt.Sprintf("Signal failed for %d: %v", pid, err)
	}

	signalName := "SIGTERM"
	if sig == syscall.SIGKILL {
		signalName = "SIGKILL"
	}
	return fmt.Sprintf("Process %d sent %s", pid, signalName)
}

func (m *Model) compactView(width, height int) string {
	lines := []string{
		fmt.Sprintf("CPU  %5.1f%%", m.cpuDisplayLoad),
		fmt.Sprintf("RAM  %5.1f%%  %s / %s", m.memDisplayPercent, ui.FormatBytes(m.memStat.Used), ui.FormatBytes(m.memStat.Total)),
	}
	if len(m.disks) > 0 {
		lines = append(lines, fmt.Sprintf("DSK  %s %5.1f%%", m.disks[0].MountPoint, m.disks[0].UsedPercent))
	}
	if len(m.gpus) > 0 {
		lines = append(lines, fmt.Sprintf("GPU  %5.1f%%", m.gpuDisplayLoad))
	}
	lines = append(lines, "Resize terminal for full dashboard")

	content := lipgloss.NewStyle().
		Width(maxInt(1, width)).
		Height(maxInt(1, height)).
		MaxWidth(maxInt(1, width)).
		MaxHeight(maxInt(1, height)).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, content)
}

func calculateLayout(width, height, diskCount int) layoutSpec {
	if width < 44 || height < 14 {
		return layoutSpec{mode: layoutCompact}
	}

	if width < 72 {
		diskHeight := minInt(ui.DiskHeight(diskCount), maxInt(5, height/3))
		remaining := maxInt(10, height-diskHeight-layoutGap)
		heights := splitWeighted(remaining, layoutGap, 2, 2, 3, 6)
		y := 0
		spec := layoutSpec{
			mode: layoutNarrow,
			cpu:  panelRect{X: 0, Y: y, W: width, H: heights[0]},
		}
		y += heights[0] + layoutGap
		spec.gpu = panelRect{X: 0, Y: y, W: width, H: heights[1]}
		y += heights[1] + layoutGap
		spec.ram = panelRect{X: 0, Y: y, W: width, H: heights[2]}
		y += heights[2] + layoutGap
		spec.disk = panelRect{X: 0, Y: y, W: width, H: diskHeight}
		y += diskHeight + layoutGap
		spec.proc = panelRect{X: 0, Y: y, W: width, H: maxInt(1, height-y)}
		return spec
	}

	widths := splitWeighted(width, layoutGap, 1, 1)
	secondRowHeight := minInt(maxInt(ui.DiskHeight(diskCount), 6), maxInt(6, height/3))
	remaining := maxInt(10, height-secondRowHeight-layoutGap)
	rowHeights := splitWeighted(remaining, layoutGap, 2, 5)
	firstRowHeight := rowHeights[0]
	procHeight := rowHeights[1]

	return layoutSpec{
		mode: layoutWide,
		cpu:  panelRect{X: 0, Y: 0, W: widths[0], H: firstRowHeight},
		gpu:  panelRect{X: widths[0] + layoutGap, Y: 0, W: widths[1], H: firstRowHeight},
		ram:  panelRect{X: 0, Y: firstRowHeight + layoutGap, W: widths[0], H: secondRowHeight},
		disk: panelRect{X: widths[0] + layoutGap, Y: firstRowHeight + layoutGap, W: widths[1], H: secondRowHeight},
		proc: panelRect{X: 0, Y: firstRowHeight + secondRowHeight + (layoutGap * 2), W: width, H: procHeight},
	}
}

func splitWeighted(total, gap int, weights ...int) []int {
	count := len(weights)
	if count == 0 {
		return nil
	}

	usable := total - gap*(count-1)
	if usable < count {
		usable = count
	}

	sumWeights := 0
	for _, w := range weights {
		sumWeights += w
	}

	sizes := make([]int, count)
	used := 0
	for i, w := range weights {
		size := usable * w / sumWeights
		if size < 1 {
			size = 1
		}
		sizes[i] = size
		used += size
	}

	for remaining, i := usable-used, 0; remaining > 0; remaining-- {
		sizes[i%count]++
		i++
	}

	for remaining, i := used-usable, 0; remaining > 0; i++ {
		idx := i % count
		if sizes[idx] > 1 {
			sizes[idx]--
			remaining--
		}
	}

	return sizes
}

func joinVerticalGap(gap int, blocks ...string) string {
	if len(blocks) == 0 {
		return ""
	}

	parts := make([]string, 0, len(blocks)*2-1)
	width := 0
	for _, block := range blocks {
		if w := lipgloss.Width(block); w > width {
			width = w
		}
	}

	for i, block := range blocks {
		if i > 0 && gap > 0 {
			parts = append(parts, lipgloss.NewStyle().Width(width).Height(gap).Render(""))
		}
		parts = append(parts, block)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func joinHorizontalGap(gap int, blocks ...string) string {
	if len(blocks) == 0 {
		return ""
	}

	parts := make([]string, 0, len(blocks)*2-1)
	height := 0
	for _, block := range blocks {
		if h := lipgloss.Height(block); h > height {
			height = h
		}
	}

	for i, block := range blocks {
		if i > 0 && gap > 0 {
			parts = append(parts, lipgloss.NewStyle().Width(gap).Height(height).Render(""))
		}
		parts = append(parts, block)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *Model) renderFooter(width int) string {
	status := m.status
	if status == "" {
		status = fmt.Sprintf("Focus: %s", m.focusLabel())
	}
	statusLine := ui.StatusStyle.Width(maxInt(1, width)).MaxWidth(maxInt(1, width)).Render(ui.Truncate(status, width))
	var help string
	switch {
	case width >= 72:
		help = strings.Join([]string{
			ui.Keycap("Tab") + ": Switch Focus",
			ui.Keycap("Arrows") + ": Navigate",
			ui.Keycap("k") + ": SigTerm",
			ui.Keycap("K") + ": SigKill",
			ui.Keycap("q") + ": Quit",
		}, " | ")
	case width >= 52:
		help = strings.Join([]string{
			ui.Keycap("Tab") + " Focus",
			ui.Keycap("↑↓") + " Move",
			ui.Keycap("k") + " Term",
			ui.Keycap("K") + " Kill",
			ui.Keycap("q") + " Quit",
		}, " | ")
	default:
		help = strings.Join([]string{
			ui.Keycap("Tab"),
			ui.Keycap("↑↓"),
			ui.Keycap("k/K"),
			ui.Keycap("q"),
		}, "  ")
	}
	helpLine := ui.FooterStyle.Width(maxInt(1, width)).MaxWidth(maxInt(1, width)).Height(1).MaxHeight(1).Render(help)
	return lipgloss.JoinVertical(lipgloss.Left, statusLine, helpLine)
}

func (m *Model) focusLabel() string {
	switch m.focused {
	case focusCPU:
		return "CPU"
	case focusGPU:
		return "GPU"
	case focusRAM:
		return "Memory"
	case focusDisk:
		return "Disks"
	case focusProcess:
		return "Processes"
	default:
		return "CPU"
	}
}

func (m *Model) advanceAnimations() {
	for i := range m.focusLevels {
		target := 0.0
		if focusPanel(i) == m.focused {
			target = 1.0
		}
		m.focusLevels[i] = easeTowards(m.focusLevels[i], target, focusAnimationStep)
	}

	m.cpuDisplayLoad = easeTowards(m.cpuDisplayLoad, m.cpuLoad, animationStep)
	m.memDisplayPercent = easeTowards(m.memDisplayPercent, m.memStat.UsedPercent, animationStep)
	m.swapDisplayPercent = easeTowards(m.swapDisplayPercent, m.memStat.SwapUsedPercent, animationStep)

	if len(m.gpus) > 0 {
		m.gpuDisplayLoad = easeTowards(m.gpuDisplayLoad, m.gpus[0].LoadPercent, animationStep)
	} else {
		m.gpuDisplayLoad = easeTowards(m.gpuDisplayLoad, 0, animationStep)
	}

	m.cpuDisplayHist = animateHistory(m.cpuDisplayHist, m.cpuHist, animationStep)
	m.memDisplayHist = animateHistory(m.memDisplayHist, m.memHist, animationStep)
	m.gpuDisplayHist = animateHistory(m.gpuDisplayHist, m.gpuHist, animationStep)
}

func easeTowards(current, target, step float64) float64 {
	if current == 0 && target == 0 {
		return 0
	}
	next := current + (target-current)*step
	if absFloat(target-next) < 0.05 {
		return target
	}
	return next
}

func animateHistory(current, target []float64, step float64) []float64 {
	if len(target) == 0 {
		return nil
	}

	if len(current) == 0 {
		return append([]float64(nil), target...)
	}

	if len(current) < len(target) {
		pad := make([]float64, len(target)-len(current))
		fill := current[len(current)-1]
		for i := range pad {
			pad[i] = fill
		}
		current = append(current, pad...)
	} else if len(current) > len(target) {
		current = current[len(current)-len(target):]
	}

	animated := make([]float64, len(target))
	for i := range target {
		animated[i] = easeTowards(current[i], target[i], step)
	}

	return animated
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
