package strategy

import (
	"fmt"
	"testing"

	"trading/model"
)

func TestMA60TrendName(t *testing.T) {
	m := &MA60Trend{}
	if m.Name() != StrategyMA60Trend {
		t.Fatalf("unexpected name: %s", m.Name())
	}
}

func TestMA60TrendScan(t *testing.T) {
	klines := buildTestKlines()
	m := &MA60Trend{}
	signals, err := m.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) == 0 {
		t.Fatal("expected at least one trend signal")
	}
	for _, s := range signals {
		if s.Strategy != StrategyMA60Trend {
			t.Fatalf("expected strategy %s, got %s", StrategyMA60Trend, s.Strategy)
		}
		if s.Score != 100 {
			t.Fatalf("expected score 100, got %f", s.Score)
		}
	}
}

func TestMA60TrendDeclining(t *testing.T) {
	klines := make([]*model.StockKline, 70)
	for i := 0; i < 70; i++ {
		price := 11.0 - float64(i)*0.01
		klines[i] = &model.StockKline{
			Code: "600312", Date: fmt.Sprintf("2026-%02d-%02d", (i/30)+1, (i%30)+1),
			Open: price, High: price + 0.1, Low: price - 0.1, Close: price, Volume: 100000,
		}
	}
	m := &MA60Trend{}
	signals, err := m.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Fatalf("expected no signals for declining trend, got %d", len(signals))
	}
}
