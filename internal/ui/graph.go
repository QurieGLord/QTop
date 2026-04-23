package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var graphBlocks = []rune("  ▂▃▄▅▆▇█")

// MultiLineGraph generates a tall history graph sized to the available card body.
func MultiLineGraph(history []float64, width, height int, color lipgloss.TerminalColor) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	if len(history) == 0 {
		empty := make([]string, 0, height)
		for i := 0; i < height; i++ {
			empty = append(empty, strings.Repeat(" ", width))
		}
		return lipgloss.JoinVertical(lipgloss.Left, empty...)
	}

	var data []float64
	if len(history) > width {
		data = history[len(history)-width:]
	} else {
		data = make([]float64, width)
		copy(data[width-len(history):], history)
	}

	var rows []string
	style := lipgloss.NewStyle().Foreground(color)

	for row := height - 1; row >= 0; row-- {
		var rowBuilder strings.Builder
		for col := 0; col < width; col++ {
			val := data[col]

			maxBlocks := float64(height * 8)
			blocks := (val / 100.0) * maxBlocks
			if blocks < 0 {
				blocks = 0
			}

			rowThreshold := float64(row * 8)

			if blocks <= rowThreshold {
				rowBuilder.WriteRune(' ')
			} else if blocks >= rowThreshold+8 {
				rowBuilder.WriteRune('█')
			} else {
				rem := blocks - rowThreshold
				idx := int(rem)
				if idx < 0 {
					idx = 0
				} else if idx >= len(graphBlocks) {
					idx = len(graphBlocks) - 1
				}
				rowBuilder.WriteRune(graphBlocks[idx])
			}
		}
		rows = append(rows, style.Render(rowBuilder.String()))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
