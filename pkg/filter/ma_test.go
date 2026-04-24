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

func TestMATrendFilterConsecutiveUp(t *testing.T) {
	klines := make([]*model.StockKline, 15)
	for i := 0; i < 15; i++ {
		klines[i] = &model.StockKline{
			Date:  fmt.Sprintf("2026-01-%02d", i+1),
			Close: 10.0 + float64(i)*0.1,
		}
	}

	f := NewMATrendUp(5).WithConsecutive(10)
	results := f.Filter(klines)

	if len(results) != 15 {
		t.Fatalf("expected 15 results, got %d", len(results))
	}

	// 前 10 天数据不足以满足连续 10 天上涨，应为 false
	for i := 0; i < 10; i++ {
		if results[i].Valid {
			t.Fatalf("day %d should not be valid when consecutive=10", i)
		}
	}

	// 第 10 天及以后，MA5 已经连续上涨 10 天
	for i := 10; i < 15; i++ {
		if !results[i].Valid {
			t.Fatalf("day %d should be valid with consecutive=10", i)
		}
	}
}

func TestMATrendFilterConsecutiveBreak(t *testing.T) {
	klines := make([]*model.StockKline, 15)
	for i := 0; i < 15; i++ {
		price := 10.0 + float64(i)*0.1
		// 第 8 天突然下跌，打断连续上涨趋势
		if i == 8 {
			price = 10.0 + float64(i-1)*0.1 - 0.5
		}
		klines[i] = &model.StockKline{
			Date:  fmt.Sprintf("2026-01-%02d", i+1),
			Close: price,
		}
	}

	f := NewMATrendUp(5).WithConsecutive(10)
	results := f.Filter(klines)

	// 第 8 天打断后，第 8~17 天都无法满足连续 10 天上涨
	for i := 8; i < 15; i++ {
		if results[i].Valid {
			t.Fatalf("day %d should not be valid after trend break at day 8", i)
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
