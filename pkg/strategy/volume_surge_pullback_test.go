package strategy

import (
	"fmt"
	"testing"

	"trading/model"
)

// buildTestKlines 构造 70 根 K 线，包含放量上涨 → 回调 → KDJ 超卖 + MA60 向上的场景
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

	// 第 30 天：放量上涨（量比 > 3，涨幅 > 3%）
	klines[30] = &model.StockKline{
		Code: "600312", Date: "2026-02-01",
		Open: 10.3, High: 10.7, Low: 10.2, Close: 10.6, Volume: 350000,
	}
	// 第 31 天：继续上涨形成峰值
	klines[31] = &model.StockKline{
		Code: "600312", Date: "2026-02-02",
		Open: 10.6, High: 11.0, Low: 10.5, Close: 10.9, Volume: 280000,
	}
	// 第 32-46 天：连续 15 天急跌，压低 KDJ
	for i := 32; i <= 46; i++ {
		prevClose := klines[i-1].Close
		close := prevClose - 0.12
		klines[i] = &model.StockKline{
			Code: "600312", Date: fmt.Sprintf("2026-02-%02d", i-29),
			Open: prevClose, High: prevClose + 0.02, Low: close - 0.02, Close: close, Volume: 60000,
		}
	}
	// 第 47-69 天：缓慢回升，保持 MA60 向上
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

// testConfig 返回测试用配置，放宽 MaxPullbackPct 和 MaxPullbackDays 以适配急跌场景
func testConfig() VolumeSurgePullbackConfig {
	cfg := DefaultVolumeSurgePullbackConfig()
	cfg.MaxPullbackPct = 20.0
	cfg.MaxPullbackDays = 15
	return cfg
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
	v := NewVolumeSurgePullback(testConfig())
	signals, err := v.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) == 0 {
		t.Fatal("expected at least one signal")
	}

	last := signals[len(signals)-1]
	if last.Type != SignalBuy {
		t.Fatalf("expected signal type buy, got %s", last.Type)
	}
	if last.Score < 70 {
		t.Fatalf("expected score >= 70, got %f", last.Score)
	}
	if last.SubScores["kdj_oversold"] <= 0 {
		t.Fatal("expected kdj_oversold score > 0")
	}
	if last.SubScores["ma60_trend"] != 100 {
		t.Fatalf("expected ma60_trend = 100, got %f", last.SubScores["ma60_trend"])
	}
}

func TestVolumeSurgePullbackNoSurge(t *testing.T) {
	// 平坦数据：没有放量上涨，不应产生任何信号
	klines := make([]*model.StockKline, 70)
	for i := 0; i < 70; i++ {
		klines[i] = &model.StockKline{
			Code: "600312", Date: fmt.Sprintf("2026-%02d-%02d", (i/30)+1, (i%30)+1),
			Open: 10.0, High: 10.1, Low: 9.9, Close: 10.0, Volume: 100000,
		}
	}
	v := NewVolumeSurgePullback(testConfig())
	signals, err := v.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Fatalf("expected no signals, got %d", len(signals))
	}
}

func TestVolumeSurgePullbackMA60Declining(t *testing.T) {
	// 整体趋势向下：MA60 一直走平/向下，即使 KDJ 超卖也不应产生信号
	klines := make([]*model.StockKline, 70)
	for i := 0; i < 70; i++ {
		price := 11.0 - float64(i)*0.01
		klines[i] = &model.StockKline{
			Code: "600312", Date: fmt.Sprintf("2026-%02d-%02d", (i/30)+1, (i%30)+1),
			Open: price + 0.05, High: price + 0.15, Low: price - 0.1, Close: price, Volume: 100000,
		}
	}
	// 在第 40 天制造一次放量上涨
	klines[40] = &model.StockKline{
		Code: "600312", Date: "2026-02-11",
		Open: 10.65, High: 10.8, Low: 10.6, Close: 10.8, Volume: 350000,
	}
	klines[41] = &model.StockKline{
		Code: "600312", Date: "2026-02-12",
		Open: 10.8, High: 11.0, Low: 10.7, Close: 10.9, Volume: 250000,
	}
	// 之后急跌
	for i := 42; i < 55; i++ {
		prevClose := klines[i-1].Close
		close := prevClose - 0.13
		klines[i] = &model.StockKline{
			Code: "600312", Date: fmt.Sprintf("2026-02-%02d", i-29),
			Open: prevClose, High: prevClose + 0.02, Low: close - 0.02, Close: close, Volume: 60000,
		}
	}

	v := NewVolumeSurgePullback(testConfig())
	signals, err := v.Scan(klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, s := range signals {
		if s.SubScores["ma60_trend"] == 100 {
			t.Fatal("expected ma60_trend = 0 when MA60 is declining")
		}
	}
}
