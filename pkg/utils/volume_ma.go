package utils

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
