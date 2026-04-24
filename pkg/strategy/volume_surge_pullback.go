package strategy

import (
	"fmt"
	"math"

	"trading/model"
	"trading/pkg/utils"
)

// VolumeSurgePullbackConfig 配置
type VolumeSurgePullbackConfig struct {
	VolumeMAPeriod  int
	MinVolumeRatio  float64
	MinRallyPct     float64
	MaxPullbackPct  float64
	MaxPullbackDays int
	KDJThreshold    float64 // J 值买入阈值，J 低于此值才触发买入
	MinScore        float64
	Weights         struct {
		VolumeSurge    float64
		PriceRally     float64
		PullbackVolume float64
		PullbackDepth  float64
		KDJOverSold    float64
		MA60Trend      float64
	}
}

// DefaultVolumeSurgePullbackConfig 返回默认配置
func DefaultVolumeSurgePullbackConfig() VolumeSurgePullbackConfig {
	return VolumeSurgePullbackConfig{
		VolumeMAPeriod:  20,
		MinVolumeRatio:  1.2,
		MinRallyPct:     2.0,
		MaxPullbackPct:  15.0,
		MaxPullbackDays: 8,
		KDJThreshold:    10.0,
		MinScore:        70.0,
		Weights: struct {
			VolumeSurge    float64
			PriceRally     float64
			PullbackVolume float64
			PullbackDepth  float64
			KDJOverSold    float64
			MA60Trend      float64
		}{
			VolumeSurge:    0.20,
			PriceRally:     0.15,
			PullbackVolume: 0.25,
			PullbackDepth:  0.20,
			KDJOverSold:    0.10,
			MA60Trend:      0.10,
		},
	}
}

// VolumeSurgePullback 放量上涨缩量回调策略
type VolumeSurgePullback struct {
	Config VolumeSurgePullbackConfig
}

// NewVolumeSurgePullback 创建策略实例
func NewVolumeSurgePullback(cfg VolumeSurgePullbackConfig) *VolumeSurgePullback {
	return &VolumeSurgePullback{Config: cfg}
}

func (v *VolumeSurgePullback) Name() string        { return StrategyVolumeSurgePullback }
func (v *VolumeSurgePullback) Description() string { return "识别放量上涨后缩量回调的技术形态" }

// DefaultConfig 返回默认配置
func (v *VolumeSurgePullback) DefaultConfig() interface{} {
	return DefaultVolumeSurgePullbackConfig()
}

// ValidateConfig 验证配置
func (v *VolumeSurgePullback) ValidateConfig(cfg interface{}) error {
	_, ok := cfg.(VolumeSurgePullbackConfig)
	if !ok {
		return fmt.Errorf("invalid config type")
	}
	return nil
}

// Scan 扫描K线序列，返回匹配的信号列表
func (v *VolumeSurgePullback) Scan(klines []*model.StockKline) ([]Signal, error) {
	cfg := v.Config
	if len(klines) < cfg.VolumeMAPeriod+1 {
		return nil, nil
	}

	volumes := make([]int64, len(klines))
	closes := make([]float64, len(klines))
	for i, k := range klines {
		volumes[i] = k.Volume
		closes[i] = k.Close
	}

	vma := utils.ComputeVolumeMA(volumes, cfg.VolumeMAPeriod)
	maMap := utils.ComputeMA(closes, []int{5, 20, 60})
	ma5 := maMap[5]
	ma20 := maMap[20]
	ma60 := maMap[60]
	kdjResults := utils.ComputeKDJ(klines)

	var signals []Signal

	for i := cfg.VolumeMAPeriod; i < len(klines); i++ {
		if vma[i] == 0 {
			continue
		}
		volRatio := float64(volumes[i]) / vma[i]
		rallyPct := (klines[i].Close - klines[i].Open) / klines[i].Open * 100
		if volRatio < cfg.MinVolumeRatio || rallyPct < cfg.MinRallyPct {
			continue
		}

		surgeIdx := i
		peakIdx := i
		peakPrice := klines[i].Close

		for j := i + 1; j < len(klines); j++ {
			currentPrice := klines[j].Close

			if currentPrice > peakPrice {
				peakIdx = j
				peakPrice = currentPrice
				continue
			}

			if currentPrice == peakPrice {
				continue
			}

			pullbackPct := (peakPrice - currentPrice) / peakPrice * 100
			pullbackDays := j - peakIdx

			if pullbackPct > cfg.MaxPullbackPct || pullbackDays > cfg.MaxPullbackDays {
				break
			}

			// 硬条件 1: KDJ J 值必须低于阈值（超卖）
			if kdjResults[j].J >= cfg.KDJThreshold {
				continue
			}

			// 硬条件 2: MA60 必须向上（当日 > 昨日）
			if j < 1 || ma60[j] <= ma60[j-1] {
				continue
			}

			score, subScores := v.calculateScore(klines, volumes, vma, ma5, ma20, ma60, kdjResults, surgeIdx, peakIdx, j)
			if score >= cfg.MinScore {
				signal := Signal{
					Code:      klines[j].Code,
					Date:      klines[j].Date,
					Strategy:  v.Name(),
					Type:      SignalBuy,
					Phase:     "pullback",
					Score:     math.Round(score*100) / 100,
					SubScores: subScores,
					Context: map[string]interface{}{
						"surge_date":       klines[surgeIdx].Date,
						"peak_date":        klines[peakIdx].Date,
						"surge_volume":     volumes[surgeIdx],
						"avg_pullback_vol": v.avgVolume(volumes, peakIdx+1, j),
						"max_pullback_pct": math.Round(pullbackPct*100) / 100,
						"pullback_days":    pullbackDays,
						"kdj_j":            math.Round(kdjResults[j].J*100) / 100,
						"ma60":             math.Round(ma60[j]*10000) / 10000,
					},
				}
				signals = append(signals, signal)
			}
		}

		if peakIdx > i {
			i = peakIdx
		}
	}

	return signals, nil
}

func (v *VolumeSurgePullback) calculateScore(
	klines []*model.StockKline,
	volumes []int64,
	vma []float64,
	_, _, ma60 []float64,
	kdjResults []utils.KDJResult,
	surgeIdx, peakIdx, currentIdx int,
) (float64, map[string]float64) {
	cfg := v.Config
	w := cfg.Weights

	volRatio := float64(volumes[surgeIdx]) / vma[surgeIdx]
	volumeSurgeScore := math.Min(volRatio, 3.0) / 3.0 * 100

	rallyPct := (klines[peakIdx].Close - klines[surgeIdx].Open) / klines[surgeIdx].Open * 100
	priceRallyScore := math.Min(rallyPct, 8.0) / 8.0 * 100

	avgPullbackVol := v.avgVolume(volumes, peakIdx+1, currentIdx)
	pullbackVolRatio := 0.0
	if volumes[surgeIdx] > 0 {
		pullbackVolRatio = float64(avgPullbackVol) / float64(volumes[surgeIdx])
	}
	pullbackVolumeScore := math.Max(0, (1.0-pullbackVolRatio)*100)

	currentPrice := klines[currentIdx].Close
	peakPrice := klines[peakIdx].Close
	pullbackPct := (peakPrice - currentPrice) / peakPrice * 100
	pullbackDepthScore := math.Max(0, (1.0-pullbackPct/cfg.MaxPullbackPct)*100)

	// KDJ 超卖得分：J 值越低分越高，J <= 0 满分
	jVal := kdjResults[currentIdx].J
	kdjScore := 0.0
	if jVal <= 0 {
		kdjScore = 100.0
	} else if jVal < cfg.KDJThreshold {
		kdjScore = (1.0 - jVal/cfg.KDJThreshold) * 100
	}

	// MA60 趋势得分：向上 = 100，走平或向下 = 0
	ma60TrendScore := 0.0
	if currentIdx >= 1 && ma60[currentIdx] > ma60[currentIdx-1] {
		ma60TrendScore = 100.0
	}

	total := volumeSurgeScore*w.VolumeSurge +
		priceRallyScore*w.PriceRally +
		pullbackVolumeScore*w.PullbackVolume +
		pullbackDepthScore*w.PullbackDepth +
		kdjScore*w.KDJOverSold +
		ma60TrendScore*w.MA60Trend

	subScores := map[string]float64{
		"volume_surge":    math.Round(volumeSurgeScore*100) / 100,
		"price_rally":     math.Round(priceRallyScore*100) / 100,
		"pullback_volume": math.Round(pullbackVolumeScore*100) / 100,
		"pullback_depth":  math.Round(pullbackDepthScore*100) / 100,
		"kdj_oversold":    math.Round(kdjScore*100) / 100,
		"ma60_trend":      math.Round(ma60TrendScore*100) / 100,
	}

	return total, subScores
}

func (v *VolumeSurgePullback) avgVolume(volumes []int64, start, end int) int64 {
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
