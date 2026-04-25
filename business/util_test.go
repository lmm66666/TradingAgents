package business

import (
	"testing"
	"time"

	"trading/model"
)

func TestToSymbol(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		want    string
		wantErr bool
	}{
		{"shanghai 6", "600312", "sh600312", false},
		{"shenzhen 0", "000001", "sz000001", false},
		{"shenzhen 3", "300001", "sz300001", false},
		{"unsupported 8", "800001", "", true},
		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toSymbol(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("toSymbol(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("toSymbol(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsFriday(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		want    bool
	}{
		{"friday", "2026-04-24", true},
		{"thursday", "2026-04-23", false},
		{"saturday", "2026-04-25", false},
		{"with time friday", "2026-04-24 15:00:00", true},
		{"with time saturday", "2026-04-25 15:00:00", false},
		{"invalid", "not-a-date", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFriday(tt.dateStr); got != tt.want {
				t.Errorf("isFriday(%q) = %v, want %v", tt.dateStr, got, tt.want)
			}
		})
	}
}

func TestFilterIncompleteWeekly(t *testing.T) {
	tests := []struct {
		name    string
		klines  []*model.StockKline
		wantLen int
	}{
		{
			name:    "empty",
			klines:  []*model.StockKline{},
			wantLen: 0,
		},
		{
			name: "last is friday",
			klines: []*model.StockKline{
				{Date: "2026-04-17", Code: "600312"},
				{Date: "2026-04-24", Code: "600312"},
			},
			wantLen: 2,
		},
		{
			name: "last is not friday",
			klines: []*model.StockKline{
				{Date: "2026-04-17", Code: "600312"},
				{Date: "2026-04-23", Code: "600312"},
			},
			wantLen: 1,
		},
		{
			name: "single friday",
			klines: []*model.StockKline{
				{Date: "2026-04-24", Code: "600312"},
			},
			wantLen: 1,
		},
		{
			name: "single not friday",
			klines: []*model.StockKline{
				{Date: "2026-04-23", Code: "600312"},
			},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterIncompleteWeekly(tt.klines)
			if len(got) != tt.wantLen {
				t.Errorf("filterIncompleteWeekly() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestFilterAfterDate(t *testing.T) {
	klines := []*model.StockKline{
		{Date: "2026-04-20", Code: "600312"},
		{Date: "2026-04-21", Code: "600312"},
		{Date: "2026-04-22", Code: "600312"},
		{Date: "2026-04-23", Code: "600312"},
		{Date: "2026-04-24", Code: "600312"},
	}

	tests := []struct {
		name     string
		lastDate string
		wantLen  int
		wantFirst string
	}{
		{"empty lastDate", "", 5, "2026-04-20"},
		{"before all", "2026-04-19", 5, "2026-04-20"},
		{"middle", "2026-04-22", 2, "2026-04-23"},
		{"last", "2026-04-24", 0, ""},
		{"after all", "2026-04-25", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterAfterDate(klines, tt.lastDate)
			if len(got) != tt.wantLen {
				t.Errorf("filterAfterDate() len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantLen > 0 && got[0].Date != tt.wantFirst {
				t.Errorf("filterAfterDate() first date = %q, want %q", got[0].Date, tt.wantFirst)
			}
		})
	}
}

func TestLastFridayDateUtil(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)

	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"friday", time.Date(2026, 4, 24, 10, 0, 0, 0, loc), "2026-04-24"},
		{"saturday", time.Date(2026, 4, 25, 10, 0, 0, 0, loc), "2026-04-24"},
		{"sunday", time.Date(2026, 4, 26, 10, 0, 0, 0, loc), "2026-04-24"},
		{"monday", time.Date(2026, 4, 27, 10, 0, 0, 0, loc), "2026-04-24"},
		{"thursday", time.Date(2026, 4, 23, 10, 0, 0, 0, loc), "2026-04-17"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lastFridayDate(tt.t); got != tt.want {
				t.Errorf("lastFridayDate(%v) = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}
