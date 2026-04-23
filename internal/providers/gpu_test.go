package providers

import (
	"errors"
	"testing"
)

func TestMockGPU(t *testing.T) {
	expected := []GPUStats{
		{
			Name:        "Mock GPU",
			LoadPercent: 42.0,
			VRAMTotal:   8 * 1024 * 1024 * 1024,
			VRAMUsed:    2 * 1024 * 1024 * 1024,
		},
	}
	mock := &MockGPU{Stats: expected}
	stats, err := mock.GetStats()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(stats) != 1 || stats[0].LoadPercent != 42.0 {
		t.Errorf("Expected LoadPercent 42.0, got %f", stats[0].LoadPercent)
	}

	mockErr := &MockGPU{Err: errors.New("gpu error")}
	_, err = mockErr.GetStats()
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

// We cannot reliably test getAmdStats or getNvidiaStats without the hardware,
// so we just test that NewGPUProvider doesn't crash.
func TestDefaultGPU(t *testing.T) {
	provider := NewGPUProvider()
	stats, err := provider.GetStats()
	// It's acceptable for this to return an empty slice and no error if no GPU is found
	if err != nil {
		t.Logf("GPU provider returned error: %v", err)
	}
	t.Logf("Found %d GPUs", len(stats))
}
