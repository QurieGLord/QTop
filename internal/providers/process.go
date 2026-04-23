package providers

import (
	"sort"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo holds information about a single process.
type ProcessInfo struct {
	PID     int32
	User    string
	Name    string
	CPU     float64
	Memory  float32
	Command string
}

// ProcessProvider defines methods to retrieve Process statistics.
type ProcessProvider interface {
	GetTopProcesses(limit int) ([]ProcessInfo, error)
}

// GopsutilProcess is an implementation of ProcessProvider.
type GopsutilProcess struct{}

// NewProcessProvider creates a new Process provider.
func NewProcessProvider() ProcessProvider {
	return &GopsutilProcess{}
}

// GetTopProcesses returns the top `limit` processes sorted by CPU usage.
func (p *GopsutilProcess) GetTopProcesses(limit int) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var infos []ProcessInfo
	for _, proc := range procs {
		cpu, err := proc.CPUPercent()
		if err != nil {
			continue // skip if we can't read cpu
		}

		// Some processes might have 0 CPU, we can optimize by only adding if > 0 or we just sort them all
		// To avoid too much overhead, we might only grab full details for top ones,
		// but we need to sort first.
		
		infos = append(infos, ProcessInfo{
			PID: proc.Pid,
			CPU: cpu,
			// Name, Memory, User, Command are fetched later for the top N to save time
		})
	}

	// Sort descending by CPU
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].CPU > infos[j].CPU
	})

	if len(infos) > limit {
		infos = infos[:limit]
	}

	// Fetch remaining details for the top N
	for i := range infos {
		proc, err := process.NewProcess(infos[i].PID)
		if err == nil {
			if name, err := proc.Name(); err == nil {
				infos[i].Name = name
			}
			if user, err := proc.Username(); err == nil {
				infos[i].User = user
			}
			if mem, err := proc.MemoryPercent(); err == nil {
				infos[i].Memory = mem
			}
			if cmd, err := proc.Cmdline(); err == nil {
				infos[i].Command = cmd
			}
		}
	}

	return infos, nil
}

// MockProcess is for testing purposes.
type MockProcess struct {
	Processes []ProcessInfo
	Err       error
}

// GetTopProcesses returns the mocked stats or error.
func (m *MockProcess) GetTopProcesses(limit int) ([]ProcessInfo, error) {
	return m.Processes, m.Err
}
