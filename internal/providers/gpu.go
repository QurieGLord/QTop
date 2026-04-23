package providers

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GPUStats holds the current GPU statistics.
type GPUStats struct {
	Name        string
	LoadPercent float64
	VRAMTotal   uint64 // in bytes
	VRAMUsed    uint64 // in bytes
}

// GPUProvider defines methods to retrieve GPU statistics.
type GPUProvider interface {
	GetStats() ([]GPUStats, error)
}

// DefaultGPU is an implementation of GPUProvider.
type DefaultGPU struct{}

// NewGPUProvider creates a new GPU provider.
func NewGPUProvider() GPUProvider {
	return &DefaultGPU{}
}

// GetStats returns stats for available GPUs.
func (p *DefaultGPU) GetStats() ([]GPUStats, error) {
	var gpus []GPUStats

	// Try NVIDIA
	nvidiaStats, err := getNvidiaStats()
	if err == nil && len(nvidiaStats) > 0 {
		gpus = append(gpus, nvidiaStats...)
	}

	// Try AMD
	amdStats, err := getAmdStats()
	if err == nil && len(amdStats) > 0 {
		gpus = append(gpus, amdStats...)
	}

	return gpus, nil
}

func getNvidiaStats() ([]GPUStats, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,utilization.gpu,memory.total,memory.used", "--format=csv,noheader,nounits")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var gpus []GPUStats
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) == 4 {
			load, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			totalMB, _ := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)
			usedMB, _ := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)

			gpus = append(gpus, GPUStats{
				Name:        strings.TrimSpace(parts[0]),
				LoadPercent: load,
				VRAMTotal:   totalMB * 1024 * 1024,
				VRAMUsed:    usedMB * 1024 * 1024,
			})
		}
	}
	return gpus, nil
}

func getAmdStats() ([]GPUStats, error) {
	var gpus []GPUStats
	matches, err := filepath.Glob("/sys/class/drm/card*/device/gpu_busy_percent")
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err == nil {
			load, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			if err == nil {
				// VRAM is trickier for AMD via sysfs, so we might just report 0 for MVP
				// We can read mem_info_vram_total and mem_info_vram_used if available
				dir := filepath.Dir(match)
				totalData, errTotal := os.ReadFile(filepath.Join(dir, "mem_info_vram_total"))
				usedData, errUsed := os.ReadFile(filepath.Join(dir, "mem_info_vram_used"))

				var total, used uint64
				if errTotal == nil && errUsed == nil {
					total, _ = strconv.ParseUint(strings.TrimSpace(string(totalData)), 10, 64)
					used, _ = strconv.ParseUint(strings.TrimSpace(string(usedData)), 10, 64)
				}

				gpus = append(gpus, GPUStats{
					Name:        "AMD GPU",
					LoadPercent: load,
					VRAMTotal:   total,
					VRAMUsed:    used,
				})
			}
		}
	}
	return gpus, nil
}

// MockGPU is for testing purposes.
type MockGPU struct {
	Stats []GPUStats
	Err   error
}

// GetStats returns the mocked stats or error.
func (m *MockGPU) GetStats() ([]GPUStats, error) {
	return m.Stats, m.Err
}
