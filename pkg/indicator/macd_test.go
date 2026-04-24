package indicator

import (
	"testing"

	"trading/model"
)

func TestComputeMACDConstant(t *testing.T) {
	klines := make([]*model.StockKline, 30)
	for i := range klines {
		klines[i] = &model.StockKline{
			Date:  "2025-01-01",
			Close: 100,
		}
	}

	results := ComputeMACD(klines)
	if len(results) != 30 {
		t.Fatalf("expected 30 results, got %d", len(results))
	}
	for i, r := range results {
		if r.DIF != 0 || r.DEA != 0 || r.BAR != 0 {
			t.Errorf("[%d] DIF=%.4f DEA=%.4f BAR=%.4f, want all 0", i, r.DIF, r.DEA, r.BAR)
		}
	}
}

func TestComputeMACDEmpty(t *testing.T) {
	results := ComputeMACD(nil)
	if results != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestComputeMACDValues(t *testing.T) {
	closes := []float64{10, 12, 11, 13, 15}
	klines := make([]*model.StockKline, len(closes))
	for i, c := range closes {
		klines[i] = &model.StockKline{Date: "2025-01-01", Close: c}
	}

	results := ComputeMACD(klines)

	if results[0].DIF != 0 {
		t.Errorf("DIF[0] = %.4f, want 0", results[0].DIF)
	}

	if results[1].DIF < 0.15 || results[1].DIF > 0.17 {
		t.Errorf("DIF[1] = %.4f, want ~0.1596", results[1].DIF)
	}
}
