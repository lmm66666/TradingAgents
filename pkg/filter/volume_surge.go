package filter

import (
	"trading/model"
	"trading/pkg/indicator"
)

type VolumeSurgeConfig struct {
	VolumeMAPeriod  int
	MinVolumeRatio  float64
	MinRallyPct     float64
	MaxPullbackPct  float64
	MaxPullbackDays int
}

func DefaultVolumeSurgeConfig() VolumeSurgeConfig {
	return VolumeSurgeConfig{
		VolumeMAPeriod:  20,
		MinVolumeRatio:  1.2,
		MinRallyPct:     2.0,
		MaxPullbackPct:  20.0,
		MaxPullbackDays: 10,
	}
}

type pullbackWindow struct {
	surgeIdx  int
	peakIdx   int
	peakPrice float64
}

type VolumeSurgeFilter struct {
	Config VolumeSurgeConfig
}

func NewVolumeSurgeFilter(cfg VolumeSurgeConfig) *VolumeSurgeFilter {
	return &VolumeSurgeFilter{Config: cfg}
}

func (v *VolumeSurgeFilter) Filter(klines []*model.StockKline) []Result {
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

	results := make([]Result, n)
	for i := range n {
		_, ok := windowByDay[i]
		results[i] = Result{
			Date:  klines[i].Date,
			Valid: ok,
		}
	}
	return results
}

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
