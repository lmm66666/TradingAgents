package score

import (
	"math"

	"trading/model"
	"trading/pkg/indicator"
)

// VolumeSurgeConfig 放量上涨+回调策略配置
type VolumeSurgeConfig struct {
	VolumeMAPeriod  int
	MinVolumeRatio  float64
	MinRallyPct     float64
	MaxPullbackPct  float64
	MaxPullbackDays int
	MinScore        float64
	Weights         struct {
		VolumeSurge    float64
		PriceRally     float64
		PullbackVolume float64
		PullbackDepth  float64
	}
}

// DefaultVolumeSurgeConfig 返回默认配置
func DefaultVolumeSurgeConfig() VolumeSurgeConfig {
	return VolumeSurgeConfig{
		VolumeMAPeriod:  20,
		MinVolumeRatio:  1.2,
		MinRallyPct:     2.0,
		MaxPullbackPct:  15.0,
		MaxPullbackDays: 8,
		MinScore:        70.0,
		Weights: struct {
			VolumeSurge    float64
			PriceRally     float64
			PullbackVolume float64
			PullbackDepth  float64
		}{
			VolumeSurge:    0.25,
			PriceRally:     0.20,
			PullbackVolume: 0.30,
			PullbackDepth:  0.25,
		},
	}
}

// pullbackWindow 记录某个回调窗口的上下文
type pullbackWindow struct {
	surgeIdx  int
	peakIdx   int
	peakPrice float64
}

// VolumeSurgeScorer 放量上涨缩量回调评分器
type VolumeSurgeScorer struct {
	Config VolumeSurgeConfig
}

// NewVolumeSurgeScorer 创建评分器实例
func NewVolumeSurgeScorer(cfg VolumeSurgeConfig) *VolumeSurgeScorer {
	return &VolumeSurgeScorer{Config: cfg}
}

// Score 返回处于放量上涨后回调区间内、评分达标的日期与分数
func (v *VolumeSurgeScorer) Score(klines []*model.StockKline) []*Result {
	cfg := v.Config
	n := len(klines)
	if n < cfg.VolumeMAPeriod+1 {
		return nil
	}

	volumes := make([]int64, n)
	for i, k := range klines {
		volumes[i] = k.Volume
	}

	vma := indicator.ComputeMA(volumes, cfg.VolumeMAPeriod)
	windows := findPullbackWindows(klines, volumes, vma, cfg.VolumeMAPeriod, cfg.MinVolumeRatio, cfg.MinRallyPct)

	// 为每个回调窗口内的天建立索引，保留最强的窗口
	windowByDay := make(map[int]*pullbackWindow)
	for i := range windows {
		w := &windows[i]
		for d := w.peakIdx + 1; d < n; d++ {
			pullbackPct := (w.peakPrice - klines[d].Close) / w.peakPrice * 100
			days := d - w.peakIdx
			if pullbackPct > cfg.MaxPullbackPct || days > cfg.MaxPullbackDays {
				break
			}
			if existing, ok := windowByDay[d]; !ok || w.peakPrice > existing.peakPrice {
				windowByDay[d] = w
			}
		}
	}

	var results []*Result
	for i := cfg.VolumeMAPeriod; i < n; i++ {
		w, ok := windowByDay[i]
		if !ok {
			continue
		}

		pullbackPct := (w.peakPrice - klines[i].Close) / w.peakPrice * 100
		score, _ := v.calculateScore(klines, volumes, vma, w, i, pullbackPct)
		if score < cfg.MinScore {
			continue
		}

		s := int(score)
		if s > 100 {
			s = 100
		}
		results = append(results, &Result{
			Date:  klines[i].Date,
			Score: s,
		})
	}

	return results
}

func (v *VolumeSurgeScorer) calculateScore(
	klines []*model.StockKline,
	volumes []int64,
	vma []float64,
	w *pullbackWindow,
	currentIdx int,
	pullbackPct float64,
) (float64, map[string]float64) {
	weights := v.Config.Weights

	volRatio := float64(volumes[w.surgeIdx]) / vma[w.surgeIdx]
	volumeSurgeScore := math.Min(volRatio, 3.0) / 3.0 * 100

	rallyPct := (klines[w.peakIdx].Close - klines[w.surgeIdx].Open) / klines[w.surgeIdx].Open * 100
	priceRallyScore := math.Min(rallyPct, 8.0) / 8.0 * 100

	pullbackVol := avgVolume(volumes, w.peakIdx+1, currentIdx)
	pullbackVolRatio := 0.0
	if volumes[w.surgeIdx] > 0 {
		pullbackVolRatio = float64(pullbackVol) / float64(volumes[w.surgeIdx])
	}
	pullbackVolumeScore := math.Max(0, (1.0-pullbackVolRatio)*100)

	pullbackDepthScore := math.Max(0, (1.0-pullbackPct/v.Config.MaxPullbackPct)*100)

	total := volumeSurgeScore*weights.VolumeSurge +
		priceRallyScore*weights.PriceRally +
		pullbackVolumeScore*weights.PullbackVolume +
		pullbackDepthScore*weights.PullbackDepth

	return total, map[string]float64{
		"volume_surge":    math.Round(volumeSurgeScore*100) / 100,
		"price_rally":     math.Round(priceRallyScore*100) / 100,
		"pullback_volume": math.Round(pullbackVolumeScore*100) / 100,
		"pullback_depth":  math.Round(pullbackDepthScore*100) / 100,
	}
}

// findPullbackWindows 找出所有放量上涨事件对应的回调窗口
func findPullbackWindows(
	klines []*model.StockKline,
	volumes []int64,
	vma []float64,
	volumeMAPeriod int,
	minVolumeRatio, minRallyPct float64,
) []pullbackWindow {
	n := len(klines)
	var windows []pullbackWindow

	for i := volumeMAPeriod; i < n; i++ {
		if vma[i] == 0 {
			continue
		}
		volRatio := float64(volumes[i]) / vma[i]
		rallyPct := (klines[i].Close - klines[i].Open) / klines[i].Open * 100
		if volRatio < minVolumeRatio || rallyPct < minRallyPct {
			continue
		}

		peakIdx := i
		peakPrice := klines[i].Close
		for j := i + 1; j < n; j++ {
			if klines[j].Close >= peakPrice {
				peakIdx = j
				peakPrice = klines[j].Close
			} else {
				break
			}
		}

		windows = append(windows, pullbackWindow{
			surgeIdx:  i,
			peakIdx:   peakIdx,
			peakPrice: peakPrice,
		})

		if peakIdx > i {
			i = peakIdx
		}
	}

	return windows
}

func avgVolume(volumes []int64, start, end int) int64 {
	if start > end || start >= len(volumes) {
		return 0
	}
	if end >= len(volumes) {
		end = len(volumes) - 1
	}
	var sum int64
	for i := start; i <= end; i++ {
		sum += volumes[i]
	}
	return sum / int64(end-start+1)
}
