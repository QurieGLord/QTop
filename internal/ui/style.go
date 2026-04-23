package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorText      = lipgloss.AdaptiveColor{Light: "0", Dark: "15"}
	ColorMuted     = lipgloss.AdaptiveColor{Light: "8", Dark: "8"}
	ColorAccent    = lipgloss.AdaptiveColor{Light: "6", Dark: "6"} // Cyan
	ColorWarning   = lipgloss.AdaptiveColor{Light: "3", Dark: "3"} // Yellow
	ColorCritical  = lipgloss.AdaptiveColor{Light: "1", Dark: "1"} // Red
	ColorSuccess   = lipgloss.AdaptiveColor{Light: "2", Dark: "2"} // Green
	ColorTitleText = lipgloss.AdaptiveColor{Light: "15", Dark: "15"}

	BaseStyle   = lipgloss.NewStyle().Foreground(ColorText)
	MutedStyle  = lipgloss.NewStyle().Foreground(ColorMuted)
	LabelStyle  = MutedStyle.Copy()
	BoldStyle   = BaseStyle.Copy().Bold(true)
	HeaderStyle = MutedStyle.Copy().Bold(true)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1).
			Foreground(ColorText)

	FocusedCardStyle = CardStyle.Copy().
				BorderForeground(ColorAccent)

	CardTitleStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorMuted).
			Bold(true).
			Padding(0, 1)

	FocusedCardTitleStyle = lipgloss.NewStyle().
				Foreground(ColorTitleText).
				Background(ColorAccent).
				Bold(true).
				Padding(0, 1)

	SelectedRowStyle = lipgloss.NewStyle().
				Background(ColorAccent).
				Foreground(ColorTitleText).
				Bold(true)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StatusStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	KeycapStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)
)

// ComponentState holds the UI state for a component.
type ComponentState struct {
	Focused bool
	Width   int
	Height  int
}

// GetColorByPercent returns a color based on the load percentage.
func GetColorByPercent(percent float64) lipgloss.TerminalColor {
	if percent >= 90 {
		return ColorCritical
	}
	if percent >= 75 {
		return ColorWarning
	}
	return ColorSuccess
}
