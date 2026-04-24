package filter

import (
	"testing"
)

func TestKDJFilterOverSold(t *testing.T) {
	klines := buildKlines(70)
	// 构造一段下跌趋势让 KDJ 走低
	for i := 10; i < 30; i++ {
		klines[i].Close = 9.0 - float64(i-10)*0.05
		klines[i].High = klines[i].Close + 0.1
		klines[i].Low = klines[i].Close - 0.1
	}

	f := NewKDJOverSold(20)
	results := f.Filter(klines)

	if len(results) == 0 {
		t.Fatal("expected results")
	}

	// 前面几天数据不足，KDJ 初始值为 50，不满足超卖
	// 后面下跌趋势中应该有满足 J < 20 的
	found := false
	for _, r := range results {
		if r.Valid {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected at least one Valid result for oversold")
	}
}

func TestKDJFilterOverBuy(t *testing.T) {
	klines := buildKlines(70)
	// 构造一段上涨趋势让 KDJ 走高
	for i := 10; i < 30; i++ {
		klines[i].Close = 11.0 + float64(i-10)*0.1
		klines[i].High = klines[i].Close + 0.1
		klines[i].Low = klines[i].Close - 0.05
	}

	f := NewKDJOverBuy(80)
	results := f.Filter(klines)

	if len(results) == 0 {
		t.Fatal("expected results")
	}

	found := false
	for _, r := range results {
		if r.Valid {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected at least one Valid result for overbuy")
	}
}

func TestKDJFilterEmpty(t *testing.T) {
	f := NewKDJOverSold(20)
	results := f.Filter(nil)
	if results != nil {
		t.Fatal("expected nil for empty klines")
	}
}
