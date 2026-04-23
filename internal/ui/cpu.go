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
	color := GetColorByPercent(m.LoadPercent)
	valueStyle := lipgloss.NewStyle().Foreground(color).Bold(true)

	rows := make([]string, 0, 3)
	if bodyH > 0 {
		rows = append(rows, metricRow("load", fmt.Sprintf("%5.1f%%", m.LoadPercent), innerW, valueStyle))
	}
	if bodyH > 1 {
		rows = append(rows, ProgressBar(innerW, m.LoadPercent, color))
	}
	if graphH := bodyH - len(rows); graphH > 0 {
		rows = append(rows, MultiLineGraph(m.History, innerW, graphH, color))
	}

	return renderCard(m.Width, m.Height, "CPU", m.Focused, lipgloss.JoinVertical(lipgloss.Left, rows...))
}
