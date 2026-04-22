package broker

import (
	"context"
	"testing"
	"time"

	"trading/model"
)

// TestSinaBrokerGetStockTodayInBatch 测试批量获取今日数据
func TestSinaBrokerGetStockTodayInBatch(t *testing.T) {
	broker := NewSinaBroker()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := broker.GetStockTodayInBatch(ctx, []string{"sh000001"})
	if err != nil {
		t.Fatalf("fetch realtime batch failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("expected at least one result, got none")
	}

	for code, kline := range result {
		t.Logf("%s: Open=%.2f High=%.2f Low=%.2f Close=%.2f Volume=%d",
			code, kline.Open, kline.High, kline.Low, kline.Close, kline.Volume)
	}
}

// TestSinaBrokerGetStockToday 测试单个获取今日数据
func TestSinaBrokerGetStockToday(t *testing.T) {
	broker := NewSinaBroker()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	kline, err := broker.GetStockToday(ctx, "sh000001")
	if err != nil {
		t.Fatalf("fetch realtime single failed: %v", err)
	}

	t.Logf("sh000001: Open=%.2f High=%.2f Low=%.2f Close=%.2f Volume=%d",
		kline.Open, kline.High, kline.Low, kline.Close, kline.Volume)
}

// TestSinaBrokerGetStockHistorical 测试获取历史K线
func TestSinaBrokerGetStockHistorical(t *testing.T) {
	broker := NewSinaBroker()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	data, err := broker.GetStockHistorical(ctx, "sh000001")
	if err != nil {
		t.Fatalf("fetch historical failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected historical data, got none")
	}

	first := data[0]
	t.Logf("First: Date=%s Open=%.2f High=%.2f Low=%.2f Close=%.2f Volume=%d",
		first.Date, first.Open, first.High, first.Low, first.Close, first.Volume)
	t.Logf("Total data points: %d", len(data))
}

// TestSinaBrokerImplementsInterface 验证 SinaBroker 实现了 IBroker 接口
func TestSinaBrokerImplementsInterface(t *testing.T) {
	var _ IBroker = (*SinaBroker)(nil)
}

// TestSinaBrokerRealtimeToKline 测试实时数据解析为 Kline
func TestSinaBrokerRealtimeToKline(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		content string
		want    *model.StockKline
	}{
		{
			name:    "stock",
			code:    "sh600000",
			content: "浦发银行,10.50,10.40,10.60,10.70,10.30,10.55,10.60,123456,130456789,f10,f11,f12,f13,f14,f15,f16,f17,f18,f19,f20,f21,f22,f23,f24,f25,f26,f27,f28,f29,2025-04-22,15:00:00,00",
			want: &model.StockKline{
				Code:   "sh600000",
				Date:   "2025-04-22",
				Open:   10.50,
				High:   10.70,
				Low:    10.30,
				Close:  10.60,
				Volume: 123456,
			},
		},
		{
			name:    "exchange",
			code:    "USDCNY",
			content: "16:30:00,7.2000,7.2100,7.2050,1000,7.2150,7.2200,7.2250,7.2100,美元人民币,2025-04-22",
			want: &model.StockKline{
				Code:   "USDCNY",
				Date:   "2025-04-22",
				Open:   7.20,
				High:   7.205,
				Low:    7.20,
				Close:  7.21,
				Volume: 1000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRealtimeToKline(tt.code, tt.content)
			if got == nil {
				t.Fatal("expected non-nil kline")
			}
			if got.Code != tt.want.Code {
				t.Errorf("Code = %s, want %s", got.Code, tt.want.Code)
			}
			if got.Date != tt.want.Date {
				t.Errorf("Date = %s, want %s", got.Date, tt.want.Date)
			}
			if got.Open != tt.want.Open {
				t.Errorf("Open = %.2f, want %.2f", got.Open, tt.want.Open)
			}
			if got.High != tt.want.High {
				t.Errorf("High = %.2f, want %.2f", got.High, tt.want.High)
			}
			if got.Low != tt.want.Low {
				t.Errorf("Low = %.2f, want %.2f", got.Low, tt.want.Low)
			}
			if got.Close != tt.want.Close {
				t.Errorf("Close = %.2f, want %.2f", got.Close, tt.want.Close)
			}
			if got.Volume != tt.want.Volume {
				t.Errorf("Volume = %d, want %d", got.Volume, tt.want.Volume)
			}
		})
	}
}
