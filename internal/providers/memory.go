package providers

import (
	"github.com/shirou/gopsutil/v3/mem"
)

type MemoryStats struct {
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64

	SwapTotal       uint64
	SwapUsed        uint64
	SwapFree        uint64
	SwapUsedPercent float64

	ZramTotal uint64
	ZramUsed  uint64
}

// MemoryProvider defines methods to retrieve Memory statistics.
type MemoryProvider interface {
	GetStats() (MemoryStats, error)
}

// GopsutilMemory is an implementation of MemoryProvider using gopsutil.
type GopsutilMemory struct{}

// NewMemoryProvider creates a new Memory provider.
func NewMemoryProvider() MemoryProvider {
	return &GopsutilMemory{}
}

func (p *GopsutilMemory) GetStats() (MemoryStats, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return MemoryStats{}, err
	}
	s, err := mem.SwapMemory()
	if err != nil {
		return MemoryStats{}, err
	}

	// For simplicity in MVP, we might leave ZRAM 0 if reading sysfs is complex.
	// But let's try to get ZRAM if we can, or just leave it 0 for now.
	return MemoryStats{
		Total:           v.Total,
		Used:            v.Used,
		Free:            v.Free,
		UsedPercent:     v.UsedPercent,
		SwapTotal:       s.Total,
		SwapUsed:        s.Used,
		SwapFree:        s.Free,
		SwapUsedPercent: s.UsedPercent,
	}, nil
}

// MockMemory is for testing purposes.
type MockMemory struct {
	Stats MemoryStats
	Err   error
}

// GetStats returns the mocked stats or error.
func (m *MockMemory) GetStats() (MemoryStats, error) {
	return m.Stats, m.Err
}
