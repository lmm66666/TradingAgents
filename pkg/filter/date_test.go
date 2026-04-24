package filter

import (
	"fmt"
	"testing"

	"trading/model"
)

func TestDateFilter(t *testing.T) {
	klines := make([]*model.StockKline, 5)
	for i := 0; i < 5; i++ {
		klines[i] = &model.StockKline{
			Date: fmt.Sprintf("2026-01-%02d", i+1),
		}
	}

	f := NewDateFilter(3)
	results := f.Filter(klines)

	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	for i, r := range results {
		if i < 3 && r.Valid {
			t.Fatalf("day %d should be false when holdDays=3", i)
		}
		if i >= 3 && !r.Valid {
			t.Fatalf("day %d should be true when holdDays=3", i)
		}
	}
}

func TestDateFilterZero(t *testing.T) {
	klines := make([]*model.StockKline, 3)
	for i := 0; i < 3; i++ {
		klines[i] = &model.StockKline{
			Date: fmt.Sprintf("2026-01-%02d", i+1),
		}
	}

	f := NewDateFilter(0)
	results := f.Filter(klines)

	for i, r := range results {
		if !r.Valid {
			t.Fatalf("day %d should be true when holdDays=0", i)
		}
	}
}

func TestDateFilterEmpty(t *testing.T) {
	f := NewDateFilter(3)
	results := f.Filter(nil)
	if results != nil {
		t.Fatal("expected nil for empty klines")
	}
}
