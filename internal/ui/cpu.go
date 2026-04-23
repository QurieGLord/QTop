package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// CPUView represents the UI component for CPU stats.
type CPUView struct {
	ComponentState
	LoadPercent float64
	History     []float64
}

// Render returns the string representation of the CPUView.
func (m CPUView) Render() string {
	innerW, bodyH := cardContentSize(m.Width, m.Height)
	density := densityForCard(innerW, bodyH)
	color := GetColorByPercent(m.LoadPercent)
	valueStyle := lipgloss.NewStyle().Foreground(color).Bold(true)

	rows := make([]string, 0, 4)
	if bodyH > 0 {
		rows = append(rows, metricRow("load", fmt.Sprintf("%5.1f%%", m.LoadPercent), innerW, valueStyle))
	}
	if density >= cardDensityCompact && bodyH > len(rows) {
		rows = append(rows, ProgressBar(innerW, m.LoadPercent, color))
	}
	if density == cardDensityCozy && bodyH > len(rows) {
		rows = append(rows, Sparkline(m.History, innerW, color))
	}
	if density == cardDensityFull {
		if graphH := bodyH - len(rows); graphH > 0 {
			rows = append(rows, MultiLineGraph(m.History, innerW, graphH, color))
		}
	} else if density == cardDensityCompact && bodyH > len(rows) {
		rows = append(rows, Sparkline(m.History, innerW, color))
	}

	return renderCard(m.Width, m.Height, "CPU", m.ComponentState, lipgloss.JoinVertical(lipgloss.Left, rows...))
}
