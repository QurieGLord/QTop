package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/QurieGLord/QTop/internal/providers"
)

func TestCPUViewRenderRespectsSize(t *testing.T) {
	view := CPUView{
		ComponentState: ComponentState{Width: 40, Height: 12, Focused: true},
		LoadPercent:    67.4,
		History:        []float64{12, 24, 36, 48, 67},
	}

	rendered := view.Render()
	if got := lipgloss.Width(rendered); got != 40 {
		t.Fatalf("expected width 40, got %d", got)
	}
	if got := lipgloss.Height(rendered); got != 12 {
		t.Fatalf("expected height 12, got %d", got)
	}
}

func TestProcessViewRenderRespectsSizeWithLongRows(t *testing.T) {
	view := ProcessView{
		ComponentState: ComponentState{Width: 72, Height: 14},
		Processes: []providers.ProcessInfo{
			{
				PID:     101,
				User:    "very-long-user-name",
				Name:    "really-long-process-name",
				CPU:     97.2,
				Memory:  13.8,
				Command: "/usr/bin/a-very-long-command --with --many --flags --and arguments",
			},
			{
				PID:     202,
				User:    "postgres",
				Name:    "postgres",
				CPU:     14.1,
				Memory:  8.3,
				Command: "postgres: checkpointer process",
			},
		},
	}

	rendered := view.Render()
	if got := lipgloss.Width(rendered); got != 72 {
		t.Fatalf("expected width 72, got %d", got)
	}
	if got := lipgloss.Height(rendered); got != 14 {
		t.Fatalf("expected height 14, got %d", got)
	}
}

func TestProgressBarRespectsRequestedWidth(t *testing.T) {
	bar := ProgressBar(18, 63.5, ColorAccent)
	if got := lipgloss.Width(bar); got != 18 {
		t.Fatalf("expected width 18, got %d", got)
	}
}

func TestDiskViewRenderRespectsSize(t *testing.T) {
	view := DiskView{
		ComponentState: ComponentState{Width: 48, Height: 10},
		Disks: []providers.DiskStats{
			{MountPoint: "/", Used: 120 * 1024 * 1024, Total: 256 * 1024 * 1024, UsedPercent: 46.9},
			{MountPoint: "/home", Used: 512 * 1024 * 1024, Total: 1024 * 1024 * 1024, UsedPercent: 50},
		},
	}

	rendered := view.Render()
	if got := lipgloss.Width(rendered); got != 48 {
		t.Fatalf("expected width 48, got %d", got)
	}
	if got := lipgloss.Height(rendered); got != 10 {
		t.Fatalf("expected height 10, got %d", got)
	}
}

func TestProcessVisibleRows(t *testing.T) {
	if got := ProcessVisibleRows(14); got <= 0 {
		t.Fatalf("expected positive visible row count, got %d", got)
	}
}

func TestProcessIndexFromY(t *testing.T) {
	index, ok := ProcessIndexFromY(14, 5, 4)
	if !ok {
		t.Fatal("expected click to resolve to process row")
	}
	if index != 6 {
		t.Fatalf("expected index 6, got %d", index)
	}
}
