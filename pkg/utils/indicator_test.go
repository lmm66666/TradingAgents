package utils

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

func TestComputeKDJBasic(t *testing.T) {
	klines := []*model.StockKline{
		{Date: "d1", High: 12, Low: 8, Close: 10},
		{Date: "d2", High: 12, Low: 8, Close: 11},
		{Date: "d3", High: 12, Low: 8, Close: 9},
	}

	results := ComputeKDJ(klines)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].K != 50 || results[0].D != 50 || results[0].J != 50 {
		t.Errorf("KDJ[0] K=%.2f D=%.2f J=%.2f, want 50/50/50", results[0].K, results[0].D, results[0].J)
	}

	if results[1].K < 58.3 || results[1].K > 58.4 {
		t.Errorf("KDJ[1].K = %.4f, want ~58.3333", results[1].K)
	}

	if results[2].K < 47.2 || results[2].K > 47.3 {
		t.Errorf("KDJ[2].K = %.4f, want ~47.2222", results[2].K)
	}
}

func TestComputeKDJNoRange(t *testing.T) {
	klines := []*model.StockKline{
		{Date: "d1", High: 10, Low: 10, Close: 10},
	}

	results := ComputeKDJ(klines)
	if results[0].K != 50 || results[0].D != 50 || results[0].J != 50 {
		t.Errorf("KDJ K=%.2f D=%.2f J=%.2f, want 50/50/50", results[0].K, results[0].D, results[0].J)
	}
}

func TestComputeKDJEmpty(t *testing.T) {
	results := ComputeKDJ(nil)
	if results != nil {
		t.Fatal("expected nil for empty input")
	}
}
