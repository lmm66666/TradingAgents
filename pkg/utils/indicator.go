package utils

import (
	"math"

	"trading/model"
)

type MACDResult struct {
	Date string  `json:"date"`
	DIF  float64 `json:"dif"`
	DEA  float64 `json:"dea"`
	BAR  float64 `json:"bar"`
}

type KDJResult struct {
	Date string  `json:"date"`
	K    float64 `json:"k"`
	D    float64 `json:"d"`
	J    float64 `json:"j"`
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
			DIF:  round4(dif[i]),
			DEA:  round4(dea[i]),
			BAR:  round4(2 * (dif[i] - dea[i])),
		}
	}
	return results
}

func calcEMA(values []float64, period int) []float64 {
	n := len(values)
	if n == 0 {
		return nil
	}

	alpha := 2.0 / float64(period+1)
	result := make([]float64, n)
	result[0] = values[0]
	for i := 1; i < n; i++ {
		result[i] = alpha*values[i] + (1-alpha)*result[i-1]
	}
	return result
}

func calcEMAFromSlice(values []float64, period int) []float64 {
	n := len(values)
	if n == 0 {
		return nil
	}

	alpha := 2.0 / float64(period+1)
	result := make([]float64, n)
	result[0] = values[0]
	for i := 1; i < n; i++ {
		effAlpha := alpha
		if i < period {
			effAlpha = 2.0 / float64(i+2)
		}
		result[i] = effAlpha*values[i] + (1-effAlpha)*result[i-1]
	}
	return result
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
			K:    round4(kVal),
			D:    round4(dVal),
			J:    round4(jVal),
		}
	}
	return results
}

// ComputeMA 计算简单移动平均（SMA）
// periods 为周期列表，如 []int{10, 20, 60}
// 返回每个周期对应的 MA 数组，键为周期
func ComputeMA(closes []float64, periods []int) map[int][]float64 {
	result := make(map[int][]float64, len(periods))
	for _, p := range periods {
		ma := make([]float64, len(closes))
		for i := range closes {
			start := 0
			if i >= p-1 {
				start = i - p + 1
			}
			sum := 0.0
			count := 0
			for j := start; j <= i; j++ {
				sum += closes[j]
				count++
			}
			ma[i] = round4(sum / float64(count))
		}
		result[p] = ma
	}
	return result
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
