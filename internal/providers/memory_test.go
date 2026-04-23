package providers

import (
	"errors"
	"testing"
)

func TestMockMemory(t *testing.T) {
	expected := MemoryStats{
		Total:       1024,
		Used:        512,
		Free:        512,
		UsedPercent: 50.0,
	}
	mock := &MockMemory{Stats: expected}
	stats, err := mock.GetStats()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if stats.UsedPercent != 50.0 {
		t.Errorf("Expected UsedPercent 50.0, got %f", stats.UsedPercent)
	}

	mockErr := &MockMemory{Err: errors.New("memory error")}
	_, err = mockErr.GetStats()
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

func TestGopsutilMemory(t *testing.T) {
	provider := NewMemoryProvider()
	stats, err := provider.GetStats()
	if err != nil {
		t.Fatalf("Failed to get memory stats: %v", err)
	}
	if stats.Total == 0 {
		t.Errorf("Total memory should be greater than 0")
	}
	if stats.UsedPercent < 0 || stats.UsedPercent > 100 {
		t.Errorf("UsedPercent out of bounds [0, 100]: %f", stats.UsedPercent)
	}
}
