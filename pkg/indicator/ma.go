package indicator

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// ComputeMA 计算简单移动平均（SMA）
func ComputeMA[T Number](values []T, period int) []float64 {
	if period <= 0 || len(values) == 0 {
		return nil
	}
	result := make([]float64, len(values))
	for i := range values {
		start := 0
		if i >= period-1 {
			start = i - period + 1
		}
		var sum float64
		count := 0
		for j := start; j <= i; j++ {
			sum += float64(values[j])
			count++
		}
		result[i] = sum / float64(count)
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
