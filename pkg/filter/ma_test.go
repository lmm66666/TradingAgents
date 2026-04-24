package filter

import (
	"fmt"
	"testing"

	"trading/model"
)

func TestMATrendFilterUp(t *testing.T) {
	klines := make([]*model.StockKline, 10)
	for i := 0; i < 10; i++ {
		klines[i] = &model.StockKline{
			Date:  fmt.Sprintf("2026-01-%02d", i+1),
			Close: 10.0 + float64(i)*0.1,
		}
	}

	f := NewMATrendUp(5)
	results := f.Filter(klines)

	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}

	// 第 0 天没有前一天对比，应为 false
	if results[0].Valid {
		t.Fatal("first day should not be valid")
	}

	// 持续上涨，MA5 应持续向上
	for i := 6; i < 10; i++ {
		if !results[i].Valid {
			t.Fatalf("day %d should be valid in uptrend", i)
		}
	}
}

func TestMATrendFilterDown(t *testing.T) {
	klines := make([]*model.StockKline, 10)
	for i := 0; i < 10; i++ {
		klines[i] = &model.StockKline{
			Date:  fmt.Sprintf("2026-01-%02d", i+1),
			Close: 10.0 - float64(i)*0.1,
		}
	}

	f := NewMATrendDown(5)
	results := f.Filter(klines)

	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}

	// 第 0 天没有前一天对比，应为 false
	if results[0].Valid {
		t.Fatal("first day should not be valid")
	}

	// 持续下跌，MA5 应持续向下
	for i := 6; i < 10; i++ {
		if !results[i].Valid {
			t.Fatalf("day %d should be valid in downtrend", i)
		}
	}
}

func TestMATrendFilterEmpty(t *testing.T) {
	f := NewMATrendUp(5)
	results := f.Filter(nil)
	if results != nil {
		t.Fatal("expected nil for empty klines")
	}
}
