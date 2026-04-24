package strategy

import (
	"fmt"
	"testing"

	"trading/model"
)

func buildTestKlines() []*model.StockKline {
	klines := make([]*model.StockKline, 25)
	for i := 0; i < 25; i++ {
		vol := int64(100000)
		close := 10.0
		open := 9.9
		high := 10.1
		low := 9.8

		switch i {
		case 20: // surge day
			vol = 300000
			open = 10.2
			close = 10.5
			high = 10.6
			low = 10.1
		case 21: // rally continues
			vol = 250000
			open = 10.5
			close = 10.8
			high = 10.9
			low = 10.4
		case 22, 23, 24: // pullback
			vol = 80000
			open = close
			close = 10.8 - float64(i-21)*0.15
			high = open + 0.1
			low = close - 0.1
		}

		klines[i] = &model.StockKline{
			Code:   "600312",
			Date:   fmt.Sprintf("2026-03-%02d", i+1),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: vol,
		}
	}
	return klines
}

func TestVolumeSurgePullbackName(t *testing.T) {
	v := NewVolumeSurgePullback(DefaultVolumeSurgePullbackConfig())
	if v.Name() != StrategyVolumeSurgePullback {
		t.Fatalf("unexpected name: %s", v.Name())
	}
	if v.Description() == "" {
		t.Fatal("description should not be empty")
	}
}

func TestVolumeSurgePullbackScan(t *testing.T) {
	klines := buildTestKlines()
	v := NewVolumeSurgePullback(DefaultVolumeSurgePullbackConfig())
	signals, err := v.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) == 0 {
		t.Fatal("expected at least one signal")
	}

	last := signals[len(signals)-1]
	if last.Phase != "pullback" {
		t.Fatalf("expected phase pullback, got %s", last.Phase)
	}
	if last.Score < 70 {
		t.Fatalf("expected score >= 70, got %f", last.Score)
	}
}

func TestVolumeSurgePullbackNoPattern(t *testing.T) {
	klines := make([]*model.StockKline, 25)
	for i := 0; i < 25; i++ {
		klines[i] = &model.StockKline{
			Code:   "600312",
			Date:   fmt.Sprintf("2026-03-%02d", i+1),
			Open:   10.0,
			High:   10.1,
			Low:    9.9,
			Close:  10.0,
			Volume: 100000,
		}
	}
	v := NewVolumeSurgePullback(DefaultVolumeSurgePullbackConfig())
	signals, err := v.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Fatalf("expected no signals, got %d", len(signals))
	}
}
