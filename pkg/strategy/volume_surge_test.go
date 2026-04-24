package strategy

import (
	"fmt"
	"testing"

	"trading/model"
)

func buildTestKlines() []*model.StockKline {
	klines := make([]*model.StockKline, 70)
	for i := 0; i < 70; i++ {
		price := 10.0 + float64(i)*0.01
		klines[i] = &model.StockKline{
			Code:   "600312",
			Date:   fmt.Sprintf("2026-%02d-%02d", (i/30)+1, (i%30)+1),
			Open:   price - 0.05,
			High:   price + 0.1,
			Low:    price - 0.1,
			Close:  price,
			Volume: 100000,
		}
	}
	klines[30] = &model.StockKline{
		Code: "600312", Date: "2026-02-01",
		Open: 10.3, High: 10.7, Low: 10.2, Close: 10.6, Volume: 350000,
	}
	klines[31] = &model.StockKline{
		Code: "600312", Date: "2026-02-02",
		Open: 10.6, High: 11.0, Low: 10.5, Close: 10.9, Volume: 280000,
	}
	for i := 32; i <= 46; i++ {
		prevClose := klines[i-1].Close
		close := prevClose - 0.12
		klines[i] = &model.StockKline{
			Code: "600312", Date: fmt.Sprintf("2026-02-%02d", i-29),
			Open: prevClose, High: prevClose + 0.02, Low: close - 0.02, Close: close, Volume: 60000,
		}
	}
	for i := 47; i < 70; i++ {
		prevClose := klines[i-1].Close
		close := prevClose + 0.03
		klines[i] = &model.StockKline{
			Code: "600312", Date: fmt.Sprintf("2026-03-%02d", i-46),
			Open: prevClose, High: close + 0.05, Low: prevClose - 0.05, Close: close, Volume: 100000,
		}
	}
	return klines
}

func testConfig() VolumeSurgeConfig {
	cfg := DefaultVolumeSurgeConfig()
	cfg.MaxPullbackPct = 20.0
	cfg.MaxPullbackDays = 15
	return cfg
}

func TestVolumeSurgeName(t *testing.T) {
	v := NewVolumeSurge(DefaultVolumeSurgeConfig())
	if v.Name() != StrategyVolumeSurge {
		t.Fatalf("unexpected name: %s", v.Name())
	}
}

func TestVolumeSurgeScan(t *testing.T) {
	klines := buildTestKlines()
	v := NewVolumeSurge(testConfig())
	signals, err := v.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) == 0 {
		t.Fatal("expected at least one signal")
	}

	for _, s := range signals {
		if s.Strategy != StrategyVolumeSurge {
			t.Fatalf("expected strategy %s, got %s", StrategyVolumeSurge, s.Strategy)
		}
		if s.Score < 70 {
			t.Fatalf("expected score >= 70, got %f", s.Score)
		}
		if _, ok := s.Context["surge_date"]; !ok {
			t.Fatal("expected surge_date in context")
		}
	}
}

func TestVolumeSurgeNoSurge(t *testing.T) {
	klines := make([]*model.StockKline, 70)
	for i := 0; i < 70; i++ {
		klines[i] = &model.StockKline{
			Code: "600312", Date: fmt.Sprintf("2026-%02d-%02d", (i/30)+1, (i%30)+1),
			Open: 10.0, High: 10.1, Low: 9.9, Close: 10.0, Volume: 100000,
		}
	}
	v := NewVolumeSurge(testConfig())
	signals, err := v.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Fatalf("expected no signals, got %d", len(signals))
	}
}
