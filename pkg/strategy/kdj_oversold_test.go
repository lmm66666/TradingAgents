package strategy

import (
	"fmt"
	"testing"

	"trading/model"
)

func TestKDJOverSoldName(t *testing.T) {
	k := NewKDJOverSold(DefaultKDJOverSoldConfig())
	if k.Name() != StrategyKDJOverSold {
		t.Fatalf("unexpected name: %s", k.Name())
	}
}

func TestKDJOverSoldScan(t *testing.T) {
	klines := buildTestKlines()
	k := NewKDJOverSold(DefaultKDJOverSoldConfig())
	signals, err := k.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) == 0 {
		t.Fatal("expected at least one oversold signal")
	}
	for _, s := range signals {
		if s.Strategy != StrategyKDJOverSold {
			t.Fatalf("expected strategy %s, got %s", StrategyKDJOverSold, s.Strategy)
		}
	}
}

func TestKDJOverSoldFlatData(t *testing.T) {
	klines := make([]*model.StockKline, 30)
	for i := 0; i < 30; i++ {
		klines[i] = &model.StockKline{
			Code: "600312", Date: fmt.Sprintf("2026-03-%02d", i+1),
			Open: 10.0, High: 10.1, Low: 9.9, Close: 10.0, Volume: 100000,
		}
	}
	k := NewKDJOverSold(DefaultKDJOverSoldConfig())
	signals, err := k.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 平坦数据 KDJ 不会超卖
	if len(signals) != 0 {
		t.Fatalf("expected no signals for flat data, got %d", len(signals))
	}
}
