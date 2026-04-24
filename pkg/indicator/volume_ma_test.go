package indicator

import (
	"testing"
)

func TestComputeVolumeMA(t *testing.T) {
	volumes := []int64{100, 200, 300, 400, 500}
	result := ComputeVolumeMA(volumes, 3)
	if len(result) != 5 {
		t.Fatalf("expected 5 results, got %d", len(result))
	}
	if result[2] != 200.0 {
		t.Fatalf("expected 200.0 at index 2, got %f", result[2])
	}
	if result[4] != 400.0 {
		t.Fatalf("expected 400.0 at index 4, got %f", result[4])
	}
}
