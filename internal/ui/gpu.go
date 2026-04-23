package ui

import (
	"fmt"

	"github.com/QurieGLord/QTop/internal/providers"
	"github.com/charmbracelet/lipgloss"
)

// GPUView represents the UI component for GPU stats.
type GPUView struct {
	ComponentState
	GPUs    []providers.GPUStats
	History []float64 // Assuming history for the first/primary GPU
}

// Render returns the string representation of the GPUView.
func (m GPUView) Render() string {
	innerW, bodyH := cardContentSize(m.Width, m.Height)
	density := densityForCard(innerW, bodyH)

	if len(m.GPUs) == 0 {
		body := ""
		if bodyH > 0 {
			body = MutedStyle.Render("No GPUs detected")
		}
		return renderCard(m.Width, m.Height, "GPU", m.ComponentState, body)
	}

	primary := m.GPUs[0]
	color := GetColorByPercent(primary.LoadPercent)
	valueStyle := lipgloss.NewStyle().Foreground(color).Bold(true)

	vram := "VRAM N/A"
	if primary.VRAMTotal > 0 {
		vram = fmt.Sprintf("%s / %s", FormatBytes(primary.VRAMUsed), FormatBytes(primary.VRAMTotal))
	}

	rows := make([]string, 0, 5)
	if bodyH > 0 {
		name := primary.Name
		if name == "" {
			name = "Primary GPU"
		}
		rows = append(rows, metricRow(truncateTextWidth(name, innerW), fmt.Sprintf("%5.1f%%", primary.LoadPercent), innerW, valueStyle))
	}
	if density >= cardDensityCozy && bodyH > len(rows) {
		rows = append(rows, metricRow("vram", vram, innerW, BoldStyle.Copy().Foreground(ColorText)))
	}
	if density >= cardDensityCompact && bodyH > len(rows) {
		rows = append(rows, ProgressBar(innerW, primary.LoadPercent, color))
	}
	if density == cardDensityCozy && bodyH > len(rows) {
		rows = append(rows, Sparkline(m.History, innerW, color))
	}
	if density == cardDensityFull {
		if graphH := bodyH - len(rows); graphH > 0 {
			rows = append(rows, MultiLineGraph(m.History, innerW, graphH, color))
		}
	}

	return renderCard(m.Width, m.Height, "GPU", m.ComponentState, lipgloss.JoinVertical(lipgloss.Left, rows...))
}
