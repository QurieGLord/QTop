package app

import (
	"testing"

	"github.com/QurieGLord/QTop/internal/providers"
)

func TestAppendHistCapsAtLimit(t *testing.T) {
	hist := make([]float64, 0, historyLimit+10)
	for i := 0; i < historyLimit+10; i++ {
		hist = appendHist(hist, float64(i))
	}

	if len(hist) != historyLimit {
		t.Fatalf("expected history length %d, got %d", historyLimit, len(hist))
	}
	if hist[0] != 10 {
		t.Fatalf("expected oldest retained value to be 10, got %.1f", hist[0])
	}
}

func TestCalculateLayoutWide(t *testing.T) {
	spec := calculateLayout(160, 48, 3)
	if spec.mode != layoutWide {
		t.Fatalf("expected wide layout, got %v", spec.mode)
	}

	checkRect := func(name string, rect panelRect) {
		if rect.W <= 0 || rect.H <= 0 {
			t.Fatalf("%s has non-positive size: %+v", name, rect)
		}
		if rect.X < 0 || rect.Y < 0 {
			t.Fatalf("%s starts outside viewport: %+v", name, rect)
		}
		if rect.X+rect.W > 160 || rect.Y+rect.H > 48 {
			t.Fatalf("%s exceeds viewport: %+v", name, rect)
		}
	}

	checkRect("cpu", spec.cpu)
	checkRect("gpu", spec.gpu)
	checkRect("ram", spec.ram)
	checkRect("disk", spec.disk)
	checkRect("proc", spec.proc)

	if spec.gpu.Y != 0 {
		t.Fatalf("expected GPU card to share the first row with CPU, got cpu=%+v gpu=%+v", spec.cpu, spec.gpu)
	}
	if spec.gpu.X != spec.cpu.W+layoutGap {
		t.Fatalf("expected GPU card to start after CPU card, got cpu=%+v gpu=%+v", spec.cpu, spec.gpu)
	}
	if spec.ram.Y != spec.cpu.H+layoutGap {
		t.Fatalf("expected memory row below cpu row, got cpu=%+v ram=%+v", spec.cpu, spec.ram)
	}
	if spec.disk.X != spec.ram.W+layoutGap || spec.disk.Y != spec.ram.Y {
		t.Fatalf("expected disks to sit beside memory, got ram=%+v disk=%+v", spec.ram, spec.disk)
	}
	if spec.proc.X != 0 || spec.proc.W != 160 {
		t.Fatalf("expected processes to span full width, got %+v", spec.proc)
	}
	if spec.proc.Y != spec.ram.Y+spec.ram.H+layoutGap {
		t.Fatalf("expected processes below second row, got ram=%+v proc=%+v", spec.ram, spec.proc)
	}
}

func TestCalculateLayoutNarrow(t *testing.T) {
	spec := calculateLayout(60, 36, 2)
	if spec.mode != layoutNarrow {
		t.Fatalf("expected narrow layout, got %v", spec.mode)
	}

	rects := []panelRect{spec.cpu, spec.gpu, spec.ram, spec.disk, spec.proc}
	y := 0
	for i, rect := range rects {
		if rect.X != 0 || rect.W != 60 {
			t.Fatalf("panel %d expected full width, got %+v", i, rect)
		}
		if rect.Y != y {
			t.Fatalf("panel %d expected y=%d, got %+v", i, y, rect)
		}
		y += rect.H
		if i < len(rects)-1 {
			y += layoutGap
		}
	}

	if y != 36 {
		t.Fatalf("expected stacked panels to consume full height, got %d", y)
	}
}

func TestCalculateLayoutCompact(t *testing.T) {
	spec := calculateLayout(40, 12, 1)
	if spec.mode != layoutCompact {
		t.Fatalf("expected compact layout, got %v", spec.mode)
	}
}

func TestAdvanceAnimationsMovesDisplayState(t *testing.T) {
	m := NewModel()
	m.cpuLoad = 90
	m.cpuHist = []float64{10, 90}
	m.memStat.UsedPercent = 75
	m.memStat.SwapUsedPercent = 20
	m.memHist = []float64{40, 75}
	m.gpus = []providers.GPUStats{{LoadPercent: 60}}
	m.gpuHist = []float64{15, 60}
	m.focused = focusGPU

	m.advanceAnimations()

	if m.cpuDisplayLoad <= 0 || m.cpuDisplayLoad >= m.cpuLoad {
		t.Fatalf("expected cpu display load to move toward target, got %.2f", m.cpuDisplayLoad)
	}
	if m.focusLevels[focusGPU] <= 0 || m.focusLevels[focusCPU] >= 1 {
		t.Fatalf("expected focus animation to move toward GPU, levels=%v", m.focusLevels)
	}
	if len(m.cpuDisplayHist) != len(m.cpuHist) {
		t.Fatalf("expected animated history length %d, got %d", len(m.cpuHist), len(m.cpuDisplayHist))
	}
}
