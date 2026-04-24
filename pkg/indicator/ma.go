package indicator

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
			ma[i] = Round4(sum / float64(count))
		}
		result[p] = ma
	}
	return result
}

// ComputeVolumeMA 计算成交量简单移动平均
func ComputeVolumeMA(volumes []int64, period int) []float64 {
	if period <= 0 || len(volumes) == 0 {
		return nil
	}
	result := make([]float64, len(volumes))
	for i := range volumes {
		start := 0
		if i >= period-1 {
			start = i - period + 1
		}
		var sum int64
		count := 0
		for j := start; j <= i; j++ {
			sum += volumes[j]
			count++
		}
		result[i] = float64(sum) / float64(count)
	}
	return result
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
