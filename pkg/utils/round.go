package utils

import "math"

// Round4 四舍五入到 4 位小数
func Round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
