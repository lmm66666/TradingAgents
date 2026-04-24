package utils

import (
	"testing"

	"trading/model"
)

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
