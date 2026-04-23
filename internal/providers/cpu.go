package providers

import (
	"github.com/shirou/gopsutil/v3/cpu"
)

// CPUProvider defines methods to retrieve CPU statistics.
type CPUProvider interface {
	GetTotalLoad() (float64, error)
}

// GopsutilCPU is an implementation of CPUProvider using gopsutil.
type GopsutilCPU struct{}

// NewCPUProvider creates a new CPU provider.
func NewCPUProvider() CPUProvider {
	return &GopsutilCPU{}
}

// GetTotalLoad returns the total CPU load percentage (0.0 - 100.0).
func (p *GopsutilCPU) GetTotalLoad() (float64, error) {
	// Fetching load without interval (0) gets the load since last call or boot.
	// For better accuracy over ticks, we could use an interval, but since we tick
	// externally, 0 is often sufficient, or we rely on the difference. 
	// cpu.Percent(0, false) usually works if called periodically.
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}
	if len(percentages) == 0 {
		return 0, nil
	}
	return percentages[0], nil
}

// MockCPU is for testing purposes.
type MockCPU struct {
	Load float64
	Err  error
}

// GetTotalLoad returns the mocked load or error.
func (m *MockCPU) GetTotalLoad() (float64, error) {
	return m.Load, m.Err
}
