package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// MemoryView represents the UI component for RAM stats.
type MemoryView struct {
	ComponentState
	Total       uint64
	Used        uint64
	UsedPercent float64

	SwapTotal       uint64
	SwapUsed        uint64
	SwapUsedPercent float64

	History []float64
}

// Render returns the string representation of the MemoryView.
func (m MemoryView) Render() string {
	innerW, bodyH := cardContentSize(m.Width, m.Height)
	density := densityForCard(innerW, bodyH)
	color := GetColorByPercent(m.UsedPercent)
	valueStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
	swapColor := GetColorByPercent(m.SwapUsedPercent)
	swapStyle := lipgloss.NewStyle().Foreground(swapColor).Bold(true)

	rows := make([]string, 0, 6)
	if bodyH > 0 {
		rows = append(rows, metricRow(
			fmt.Sprintf("ram  %s / %s", FormatBytes(m.Used), FormatBytes(m.Total)),
			fmt.Sprintf("%5.1f%%", m.UsedPercent),
			innerW,
			valueStyle,
		))
	}
	if density >= cardDensityCompact && bodyH > len(rows) {
		rows = append(rows, ProgressBar(innerW, m.UsedPercent, color))
	}
	if density >= cardDensityCozy && bodyH > len(rows) {
		rows = append(rows, metricRow(
			fmt.Sprintf("swap %s / %s", FormatBytes(m.SwapUsed), FormatBytes(m.SwapTotal)),
			fmt.Sprintf("%5.1f%%", m.SwapUsedPercent),
			innerW,
			swapStyle,
		))
	}
	if density == cardDensityFull && bodyH > len(rows) {
		rows = append(rows, ProgressBar(innerW, m.SwapUsedPercent, swapColor))
	}
	if density == cardDensityCozy && bodyH > len(rows) {
		rows = append(rows, Sparkline(m.History, innerW, color))
	}
	if density == cardDensityFull {
		if graphH := bodyH - len(rows); graphH > 0 {
			rows = append(rows, MultiLineGraph(m.History, innerW, graphH, color))
		}
	}

	return renderCard(m.Width, m.Height, "Memory", m.ComponentState, lipgloss.JoinVertical(lipgloss.Left, rows...))
}
