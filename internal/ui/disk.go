package ui

import (
	"fmt"

	"github.com/QurieGLord/QTop/internal/providers"
	"github.com/charmbracelet/lipgloss"
)

// DiskView renders mounted storage devices.
type DiskView struct {
	ComponentState
	Disks []providers.DiskStats
}

// Render returns the string representation of the DiskView.
func (m DiskView) Render() string {
	innerW, bodyH := cardContentSize(m.Width, m.Height)
	density := densityForCard(innerW, bodyH)
	if bodyH <= 0 {
		return renderCard(m.Width, m.Height, "Disks", m.ComponentState, "")
	}

	if len(m.Disks) == 0 {
		return renderCard(m.Width, m.Height, "Disks", m.ComponentState, MutedStyle.Render("No mounted disks"))
	}

	rows := make([]string, 0, bodyH)
	diskRows := DiskRowsForHeight(m.Height)
	if diskRows <= 0 {
		diskRows = 1
	}

	for i, disk := range m.Disks {
		if len(rows) >= bodyH {
			break
		}
		if i >= diskRows {
			rows = append(rows, MutedStyle.Render(truncateTextWidth(fmt.Sprintf("+%d more", len(m.Disks)-i), innerW)))
			break
		}

		rows = append(rows, metricRow(
			disk.MountPoint,
			fmt.Sprintf("%s / %s", FormatBytes(disk.Used), FormatBytes(disk.Total)),
			innerW,
			BoldStyle.Copy().Foreground(ColorText),
		))

		if density >= cardDensityCompact && len(rows) < bodyH {
			rows = append(rows, ProgressBar(innerW, disk.UsedPercent, GetColorByPercent(disk.UsedPercent)))
		}
	}

	return renderCard(m.Width, m.Height, "Disks", m.ComponentState, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// DiskHeight returns the panel height needed for the number of disks.
func DiskHeight(count int) int {
	rows := maxInt(1, count)
	bodyRows := rows*2 + 1
	return bodyRows + CardStyle.GetVerticalFrameSize() + 1
}

// DiskRowsForHeight returns the number of disks that fit in the given panel height.
func DiskRowsForHeight(height int) int {
	_, bodyH := cardContentSize(40, height)
	if bodyH <= 0 {
		return 0
	}
	return maxInt(1, bodyH/2)
}
