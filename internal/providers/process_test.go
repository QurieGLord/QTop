package providers

import (
	"errors"
	"testing"
)

func TestMockProcess(t *testing.T) {
	expected := []ProcessInfo{
		{
			PID:     1,
			User:    "root",
			Name:    "systemd",
			CPU:     0.1,
			Memory:  0.2,
			Command: "/sbin/init",
		},
	}
	mock := &MockProcess{Processes: expected}
	procs, err := mock.GetTopProcesses(10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(procs) != 1 || procs[0].Name != "systemd" {
		t.Errorf("Expected systemd process, got %+v", procs)
	}

	mockErr := &MockProcess{Err: errors.New("proc error")}
	_, err = mockErr.GetTopProcesses(10)
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

func TestGopsutilProcess(t *testing.T) {
	provider := NewProcessProvider()
	procs, err := provider.GetTopProcesses(5)
	if err != nil {
		t.Fatalf("Failed to get processes: %v", err)
	}
	if len(procs) > 5 {
		t.Errorf("Expected at most 5 processes, got %d", len(procs))
	}
}
