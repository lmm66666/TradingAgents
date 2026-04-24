package strategy

import (
	"testing"

	"trading/model"
)

type mockStrategy struct {
	name    string
	signals []Signal
}

func (m *mockStrategy) Name() string        { return m.name }
func (m *mockStrategy) Description() string { return "mock" }
func (m *mockStrategy) Scan(klines []*model.StockKline) ([]Signal, error) {
	return m.signals, nil
}

func TestScannerScan(t *testing.T) {
	s1 := &mockStrategy{name: "s1", signals: []Signal{{Code: "600312", Score: 80}}}
	s2 := &mockStrategy{name: "s2", signals: []Signal{{Code: "000001", Score: 60}}}
	scanner := NewScanner(s1, s2)

	result, err := scanner.Scan(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 strategy results, got %d", len(result))
	}
	if len(result["s1"]) != 1 || result["s1"][0].Score != 80 {
		t.Fatalf("unexpected s1 result")
	}
	if len(result["s2"]) != 1 || result["s2"][0].Score != 60 {
		t.Fatalf("unexpected s2 result")
	}
}
