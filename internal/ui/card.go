package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

const (
	minCardWidth  = 8
	minCardHeight = 4
	cardBorderW   = 2
	cardBorderH   = 2
)

type cardDensity int

const (
	cardDensityTiny cardDensity = iota
	cardDensityCompact
	cardDensityCozy
	cardDensityFull
)

func cardContentSize(width, height int) (int, int) {
	frameW := CardStyle.GetHorizontalFrameSize()
	frameH := CardStyle.GetVerticalFrameSize()

	innerW := maxInt(1, width-frameW)
	innerH := maxInt(1, height-frameH)
	bodyH := maxInt(0, innerH-1)

	return innerW, bodyH
}

func renderCard(width, height int, title string, state ComponentState, body string) string {
	width = maxInt(1, width)
	height = maxInt(1, height)

	if width < minCardWidth || height < minCardHeight {
		compact := lipgloss.NewStyle().
			Width(width).
			Height(height).
			MaxWidth(width).
			MaxHeight(height)
		return compact.Render(truncateTextWidth(strings.ToUpper(title), width))
	}

	style := cardStyleForLevel(state.FocusLevel)

	innerW, bodyH := cardContentSize(width, height)
	titleLine := renderCardTitle(strings.ToUpper(title), innerW, state.FocusLevel)

	bodyBlock := ""
	if bodyH > 0 {
		bodyBlock = lipgloss.NewStyle().
			Width(innerW).
			Height(bodyH).
			MaxWidth(innerW).
			MaxHeight(bodyH).
			Render(body)
	}

	content := titleLine
	if bodyBlock != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, titleLine, bodyBlock)
	}

	return style.
		Width(maxInt(1, width-cardBorderW)).
		Height(maxInt(1, height-cardBorderH)).
		MaxWidth(width).
		MaxHeight(height).
		Render(content)
}

func renderCardTitle(title string, width int, focusLevel float64) string {
	style := CardTitleStyle
	prefix := "○ "
	switch focusFrame(focusLevel) {
	case 1:
		style = SoftCardTitleStyle
		prefix = "◔ "
	case 2:
		style = SoftCardTitleStyle
		prefix = "◑ "
	case 3:
		style = FocusedCardTitleStyle
		prefix = "● "
	}

	labelWidth := maxInt(1, width-style.GetHorizontalFrameSize())
	label := truncateTextWidth(prefix+title, labelWidth)

	return lipgloss.PlaceHorizontal(width, lipgloss.Left, style.Render(label))
}

func metricRow(label, value string, width int, valueStyle lipgloss.Style) string {
	if width <= 0 {
		return ""
	}

	label = truncateTextWidth(label, width)
	value = truncateTextWidth(value, width)

	gap := width - lipgloss.Width(label) - lipgloss.Width(value)
	if gap < 1 {
		maxLabelWidth := maxInt(1, width-lipgloss.Width(value)-1)
		label = truncateTextWidth(label, maxLabelWidth)
		gap = width - lipgloss.Width(label) - lipgloss.Width(value)
	}
	if gap < 1 {
		value = truncateTextWidth(value, maxInt(1, width-lipgloss.Width(label)))
		gap = maxInt(0, width-lipgloss.Width(label)-lipgloss.Width(value))
	}

	return LabelStyle.Render(label) + strings.Repeat(" ", gap) + valueStyle.Render(value)
}

func fitCell(text string, width int, alignRight bool) string {
	if width <= 0 {
		return ""
	}

	text = truncateTextWidth(text, width)
	padding := width - lipgloss.Width(text)
	if padding <= 0 {
		return text
	}
	if alignRight {
		return strings.Repeat(" ", padding) + text
	}
	return text + strings.Repeat(" ", padding)
}

func truncateTextWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}

	trimmed := text
	for lipgloss.Width(trimmed) > width-1 && len(trimmed) > 0 {
		_, size := utf8.DecodeLastRuneInString(trimmed)
		trimmed = trimmed[:len(trimmed)-size]
	}

	return trimmed + "…"
}

// Truncate trims text to the given display width.
func Truncate(text string, width int) string {
	return truncateTextWidth(text, width)
}

// Keycap renders a highlighted footer key label.
func Keycap(label string) string {
	return KeycapStyle.Render(label)
}

func focusFrame(level float64) int {
	level = clamp01(level)
	switch {
	case level >= 0.85:
		return 3
	case level >= 0.5:
		return 2
	case level >= 0.2:
		return 1
	default:
		return 0
	}
}

func cardStyleForLevel(level float64) lipgloss.Style {
	switch focusFrame(level) {
	case 1, 2:
		return SoftCardStyle
	case 3:
		return FocusedCardStyle
	default:
		return CardStyle
	}
}

func densityForCard(width, bodyH int) cardDensity {
	switch {
	case bodyH <= 1 || width < 18:
		return cardDensityTiny
	case bodyH <= 3 || width < 26:
		return cardDensityCompact
	case bodyH <= 6 || width < 36:
		return cardDensityCozy
	default:
		return cardDensityFull
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func maxInt(values ...int) int {
	best := values[0]
	for _, v := range values[1:] {
		if v > best {
			best = v
		}
	}
	return best
}

func minInt(values ...int) int {
	best := values[0]
	for _, v := range values[1:] {
		if v < best {
			best = v
		}
	}
	return best
}
