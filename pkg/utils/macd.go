package utils

import "trading/model"

type MACDResult struct {
	Date string  `json:"date"`
	DIF  float64 `json:"dif"`
	DEA  float64 `json:"dea"`
	BAR  float64 `json:"bar"`
}

// ComputeMACD 标准 MACD: EMA12, EMA26, DEA(Signal=9)
// 初始 EMA 用 close[0] 启动
func ComputeMACD(klines []*model.StockKline) []MACDResult {
	n := len(klines)
	if n == 0 {
		return nil
	}

	const (
		shortPeriod  = 12
		longPeriod   = 26
		signalPeriod = 9
	)

	closes := make([]float64, n)
	for i, k := range klines {
		closes[i] = k.Close
	}

	emaShort := calcEMA(closes, shortPeriod)
	emaLong := calcEMA(closes, longPeriod)

	dif := make([]float64, n)
	for i := range dif {
		dif[i] = emaShort[i] - emaLong[i]
	}

	dea := calcEMAFromSlice(dif, signalPeriod)

	results := make([]MACDResult, n)
	for i := range results {
		results[i] = MACDResult{
			Date: klines[i].Date,
			DIF:  Round4(dif[i]),
			DEA:  Round4(dea[i]),
			BAR:  Round4(2 * (dif[i] - dea[i])),
		}
	}
	return results
}
