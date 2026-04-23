package app

import "testing"

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

	if spec.gpu.Y != spec.cpu.H+layoutGap {
		t.Fatalf("expected GPU card below CPU card, got cpu=%+v gpu=%+v", spec.cpu, spec.gpu)
	}
	if spec.ram.X != spec.cpu.W+layoutGap {
		t.Fatalf("expected RAM card to start after CPU column, got cpu=%+v ram=%+v", spec.cpu, spec.ram)
	}
	if spec.disk.Y != spec.ram.H+layoutGap {
		t.Fatalf("expected disks below memory card, got ram=%+v disk=%+v", spec.ram, spec.disk)
	}
	if spec.proc.Y != spec.disk.Y+spec.disk.H+layoutGap {
		t.Fatalf("expected processes below disk card, got disk=%+v proc=%+v", spec.disk, spec.proc)
	}
}

func TestCalculateLayoutNarrow(t *testing.T) {
	spec := calculateLayout(92, 36, 2)
	if spec.mode != layoutNarrow {
		t.Fatalf("expected narrow layout, got %v", spec.mode)
	}

	rects := []panelRect{spec.cpu, spec.gpu, spec.ram, spec.disk, spec.proc}
	y := 0
	for i, rect := range rects {
		if rect.X != 0 || rect.W != 92 {
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
