package ui

import (
	"fmt"

	"github.com/QurieGLord/QTop/internal/providers"
	"github.com/charmbracelet/lipgloss"
)

// ProcessView represents the UI component for the process list.
type ProcessView struct {
	ComponentState
	Processes []providers.ProcessInfo
	Scroll    int
	Selected  int
}

// Render returns the string representation of the ProcessView.
func (m ProcessView) Render() string {
	innerW, bodyH := cardContentSize(m.Width, m.Height)
	if bodyH <= 0 {
		return renderCard(m.Width, m.Height, "Processes", m.ComponentState, "")
	}

	if innerW < 24 {
		lines := make([]string, 0, bodyH)
		for i := 0; i < minInt(bodyH, len(m.Processes)); i++ {
			p := m.Processes[i]
			lines = append(lines, truncateTextWidth(fmt.Sprintf("%d %s", p.PID, p.Name), innerW))
		}
		return renderCard(m.Width, m.Height, "Processes", m.ComponentState, lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	prefixW := 2
	pidW, userW, cpuW, memW, cmdW := processColumnWidths(innerW - prefixW)
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		fitCell("", prefixW, false),
		HeaderStyle.Render(fitCell("PID", pidW, true)),
		" ",
		HeaderStyle.Render(fitCell("USER", userW, false)),
		" ",
		HeaderStyle.Render(fitCell("CPU%", cpuW, true)),
		" ",
		HeaderStyle.Render(fitCell("MEM%", memW, true)),
		" ",
		HeaderStyle.Render(fitCell("COMMAND", cmdW, false)),
	)

	rows := []string{header}
	maxRows := ProcessVisibleRows(m.Height)
	start := maxInt(0, minInt(m.Scroll, maxInt(0, len(m.Processes)-maxRows)))
	end := minInt(len(m.Processes), start+maxRows)

	for i := start; i < end; i++ {
		p := m.Processes[i]
		command := p.Command
		if command == "" {
			command = p.Name
		}

		rows = append(rows, processRow(
			pidW,
			userW,
			cpuW,
			memW,
			cmdW,
			fmt.Sprintf("%d", p.PID),
			p.User,
			fmt.Sprintf("%.1f%%", p.CPU),
			fmt.Sprintf("%.1f%%", p.Memory),
			command,
			i == m.Selected,
		))
	}

	return renderCard(m.Width, m.Height, "Processes", m.ComponentState, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// ProcessVisibleRows returns the number of process rows that fit below the header.
func ProcessVisibleRows(height int) int {
	_, bodyH := cardContentSize(40, height)
	if bodyH <= 1 {
		return 0
	}
	return bodyH - 1
}

// ProcessIndexFromY resolves a click Y position within the processes card to an absolute index.
func ProcessIndexFromY(height, scroll, relativeY int) (int, bool) {
	if relativeY < 3 {
		return 0, false
	}

	index := scroll + relativeY - 3
	if index < 0 {
		return 0, false
	}
	if index >= scroll+ProcessVisibleRows(height) {
		return 0, false
	}
	return index, true
}

func processColumnWidths(total int) (int, int, int, int, int) {
	separators := 4
	usable := maxInt(1, total-separators)

	pidW := 7
	userW := minInt(14, maxInt(8, usable/6))
	cpuW := 6
	memW := 6
	cmdW := usable - pidW - userW - cpuW - memW

	if cmdW < 8 {
		shrink := 8 - cmdW
		reduceUser := minInt(shrink, maxInt(0, userW-6))
		userW -= reduceUser
		shrink -= reduceUser

		reducePID := minInt(shrink, maxInt(0, pidW-5))
		pidW -= reducePID
		shrink -= reducePID

		cmdW += reduceUser + reducePID - shrink
	}

	cmdW = maxInt(6, cmdW)

	return pidW, userW, cpuW, memW, cmdW
}

func processRow(pidW, userW, cpuW, memW, cmdW int, pid, user, cpu, mem, cmd string, selected bool) string {
	base := BaseStyle
	if selected {
		base = SelectedRowStyle
	}
	cpuStyle := base.Copy().Foreground(GetColorByPercent(parsePercent(cpu)))
	memStyle := base.Copy().Foreground(ColorAccent)

	cells := []string{
		base.Render(fitCell(pid, pidW, true)),
		base.Render(fitCell(user, userW, false)),
		cpuStyle.Render(fitCell(cpu, cpuW, true)),
		memStyle.Render(fitCell(mem, memW, true)),
		base.Render(fitCell(cmd, cmdW, false)),
	}

	row := lipgloss.JoinHorizontal(lipgloss.Left, cells[0], " ", cells[1], " ", cells[2], " ", cells[3], " ", cells[4])
	prefix := "  "
	if selected {
		prefix = "› "
	}
	return base.Render(prefix + row)
}

func parsePercent(value string) float64 {
	var percent float64
	fmt.Sscanf(value, "%f%%", &percent)
	return percent
}
