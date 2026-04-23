package providers

import (
	"sort"
	"strings"

	godisk "github.com/shirou/gopsutil/v3/disk"
)

// DiskStats holds usage information for a mounted filesystem.
type DiskStats struct {
	MountPoint  string
	FSType      string
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

// DiskProvider defines methods to retrieve disk usage.
type DiskProvider interface {
	GetStats() ([]DiskStats, error)
}

// GopsutilDisk is a gopsutil-backed disk provider.
type GopsutilDisk struct{}

// NewDiskProvider creates a new disk provider.
func NewDiskProvider() DiskProvider {
	return &GopsutilDisk{}
}

// GetStats returns the usage for relevant mounted partitions.
func (p *GopsutilDisk) GetStats() ([]DiskStats, error) {
	partitions, err := godisk.Partitions(false)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(partitions))
	stats := make([]DiskStats, 0, len(partitions))
	for _, partition := range partitions {
		if _, ok := seen[partition.Mountpoint]; ok {
			continue
		}
		if shouldSkipPartition(partition) {
			continue
		}

		usage, err := godisk.Usage(partition.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		seen[partition.Mountpoint] = struct{}{}
		stats = append(stats, DiskStats{
			MountPoint:  partition.Mountpoint,
			FSType:      partition.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return diskSortKey(stats[i].MountPoint) < diskSortKey(stats[j].MountPoint)
	})

	return stats, nil
}

func shouldSkipPartition(partition godisk.PartitionStat) bool {
	if partition.Mountpoint == "" {
		return true
	}

	for _, opt := range partition.Opts {
		if strings.Contains(opt, "loop") {
			return true
		}
	}

	switch partition.Fstype {
	case "", "autofs", "binfmt_misc", "bpf", "cgroup", "cgroup2", "configfs", "debugfs", "devpts", "devtmpfs",
		"efivarfs", "fusectl", "hugetlbfs", "mqueue", "nsfs", "overlay", "proc", "pstore", "securityfs",
		"selinuxfs", "squashfs", "sysfs", "tmpfs", "tracefs":
		return true
	}

	return false
}

func diskSortKey(mountPoint string) string {
	if mountPoint == "/" {
		return "0"
	}
	return "1" + mountPoint
}

// MockDisk is for tests.
type MockDisk struct {
	Stats []DiskStats
	Err   error
}

// GetStats returns mocked disk stats.
func (m *MockDisk) GetStats() ([]DiskStats, error) {
	return m.Stats, m.Err
}
