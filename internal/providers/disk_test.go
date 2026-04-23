package providers

import (
	"errors"
	"testing"
)

func TestMockDisk(t *testing.T) {
	expected := []DiskStats{
		{
			MountPoint:  "/",
			FSType:      "ext4",
			Total:       1024,
			Used:        512,
			Free:        512,
			UsedPercent: 50,
		},
	}

	mock := &MockDisk{Stats: expected}
	stats, err := mock.GetStats()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(stats) != 1 || stats[0].MountPoint != "/" {
		t.Fatalf("unexpected disk stats: %+v", stats)
	}

	mockErr := &MockDisk{Err: errors.New("disk error")}
	if _, err := mockErr.GetStats(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewDiskProvider(t *testing.T) {
	provider := NewDiskProvider()
	stats, err := provider.GetStats()
	if err != nil {
		t.Logf("disk provider returned error: %v", err)
	}
	t.Logf("found %d disks", len(stats))
}
