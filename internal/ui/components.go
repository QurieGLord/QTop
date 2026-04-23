package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var partialBlocks = []rune(" ▏▎▍▌▋▊▉")

// ProgressBar generates a thin progress bar with partial blocks and a muted rail.
func ProgressBar(width int, percent float64, color lipgloss.TerminalColor) string {
	if width <= 0 {
		return ""
	}

	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	total := (percent / 100.0) * float64(width)
	full := int(math.Floor(total))
	remainder := total - float64(full)
	partial := int(math.Round(remainder * float64(len(partialBlocks)-1)))

	filledStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := lipgloss.NewStyle().Foreground(ColorMuted)

	fill := strings.Repeat("█", full)
	if full < width && partial > 0 {
		fill += string(partialBlocks[partial])
		full++
	}
	rail := strings.Repeat("─", maxInt(0, width-full))

	return fmt.Sprintf("%s%s", filledStyle.Render(fill), emptyStyle.Render(rail))
}

var sparkChars = []rune(" ▂▃▄▅▆▇█")

// Sparkline generates a small graph using Unicode bar characters.
func Sparkline(history []float64, width int, color lipgloss.TerminalColor) string {
	if len(history) == 0 || width <= 0 {
		return ""
	}

	// Truncate or pad history to match width
	var data []float64
	if len(history) > width {
		data = history[len(history)-width:]
	} else {
		data = history
	}

	var builder strings.Builder
	style := lipgloss.NewStyle().Foreground(color)

	for _, val := range data {
		// Normalize 0-100 to index 0-7
		idx := int((val / 100.0) * float64(len(sparkChars)-1))
		if idx < 0 {
			idx = 0
		} else if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		builder.WriteRune(sparkChars[idx])
	}

	return style.Render(builder.String())
}

// FormatBytes formats a byte count into a human-readable string.
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
