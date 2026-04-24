package utils

import "trading/model"

type KDJResult struct {
	Date string  `json:"date"`
	K    float64 `json:"k"`
	D    float64 `json:"d"`
	J    float64 `json:"j"`
}

// ComputeKDJ 标准 KDJ: RSV(9), K/D 初始 50
func ComputeKDJ(klines []*model.StockKline) []KDJResult {
	n := len(klines)
	if n == 0 {
		return nil
	}

	const period = 9

	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)
	for i, k := range klines {
		highs[i] = k.High
		lows[i] = k.Low
		closes[i] = k.Close
	}

	results := make([]KDJResult, n)
	kVal, dVal := 50.0, 50.0

	for i := range klines {
		start := 0
		if i >= period {
			start = i - period + 1
		}

		highMax := highs[start]
		lowMin := lows[start]
		for j := start + 1; j <= i; j++ {
			if highs[j] > highMax {
				highMax = highs[j]
			}
			if lows[j] < lowMin {
				lowMin = lows[j]
			}
		}

		rsv := 50.0
		if highMax != lowMin {
			rsv = (closes[i] - lowMin) / (highMax - lowMin) * 100
		}

		kVal = 2.0/3*kVal + 1.0/3*rsv
		dVal = 2.0/3*dVal + 1.0/3*kVal
		jVal := 3*kVal - 2*dVal

		results[i] = KDJResult{
			Date: klines[i].Date,
			K:    Round4(kVal),
			D:    Round4(dVal),
			J:    Round4(jVal),
		}
	}
	return results
}
