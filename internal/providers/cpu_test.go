package providers

import (
	"errors"
	"testing"
)

func TestMockCPU(t *testing.T) {
	mock := &MockCPU{Load: 42.5}
	load, err := mock.GetTotalLoad()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if load != 42.5 {
		t.Errorf("Expected load 42.5, got %f", load)
	}

	mockErr := &MockCPU{Err: errors.New("cpu error")}
	_, err = mockErr.GetTotalLoad()
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

func TestGopsutilCPU(t *testing.T) {
	provider := NewCPUProvider()
	load, err := provider.GetTotalLoad()
	if err != nil {
		t.Fatalf("Failed to get total load: %v", err)
	}
	if load < 0 || load > 100 {
		t.Errorf("Load out of bounds [0, 100]: %f", load)
	}
}
