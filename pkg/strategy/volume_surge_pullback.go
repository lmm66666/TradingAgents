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
	surgeIdx       int
	peakIdx        int
	peakPrice      float64
	pullbackPct    float64
	pullbackDays   int
	avgPullbackVol int64
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
// 三个条件独立计算，取交集：
//  1. 处于放量上涨后的有效回调窗口内
//  2. KDJ J 值 < 阈值（超卖）
//  3. MA60 向上（当日 > 昨日）
func (v *VolumeSurgePullback) Scan(klines []*model.StockKline) ([]Signal, error) {
	cfg := v.Config
	n := len(klines)
	if n < cfg.VolumeMAPeriod+1 {
		return nil, nil
	}

	// Step 1: 预计算所有指标
	volumes := make([]int64, n)
	closes := make([]float64, n)
	for i, k := range klines {
		volumes[i] = k.Volume
		closes[i] = k.Close
	}

	vma := utils.ComputeVolumeMA(volumes, cfg.VolumeMAPeriod)
	ma60 := utils.ComputeMA(closes, []int{60})[60]
	kdjResults := utils.ComputeKDJ(klines)

	// Step 2: 找出所有放量上涨+回调窗口
	pullbackWindows := v.findPullbackWindows(klines, volumes, vma)

	// 为每个回调窗口内的天建立索引，保留最强的窗口
	windowByDay := make(map[int]*pullbackWindow)
	for i := range pullbackWindows {
		w := &pullbackWindows[i]
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

	// Step 3: 找出所有 KDJ J < 阈值的天
	kdjOversold := make(map[int]bool)
	for i := range kdjResults {
		if kdjResults[i].J < cfg.KDJThreshold {
			kdjOversold[i] = true
		}
	}

	// Step 4: 找出所有 MA60 向上的天
	ma60Rising := make(map[int]bool)
	for i := 1; i < n; i++ {
		if ma60[i] > ma60[i-1] {
			ma60Rising[i] = true
		}
	}

	// Step 5: 三个条件取交集 → 生成 Buy 信号
	var signals []Signal
	for i := cfg.VolumeMAPeriod; i < n; i++ {
		w, inWindow := windowByDay[i]
		if !inWindow || !kdjOversold[i] || !ma60Rising[i] {
			continue
		}

		score, subScores := v.calculateScore(klines, volumes, vma, w, i)
		if score < cfg.MinScore {
			continue
		}

		signal := Signal{
			Code:      klines[i].Code,
			Date:      klines[i].Date,
			Strategy:  v.Name(),
			Type:      SignalBuy,
			Phase:     "pullback",
			Score:     math.Round(score*100) / 100,
			SubScores: subScores,
			Context: map[string]interface{}{
				"surge_date":       klines[w.surgeIdx].Date,
				"peak_date":        klines[w.peakIdx].Date,
				"surge_volume":     volumes[w.surgeIdx],
				"avg_pullback_vol": w.avgPullbackVol,
				"max_pullback_pct": math.Round(w.pullbackPct*100) / 100,
				"pullback_days":    w.pullbackDays,
				"kdj_j":            math.Round(kdjResults[i].J*100) / 100,
				"ma60":             math.Round(ma60[i]*10000) / 10000,
			},
		}
		signals = append(signals, signal)
	}

	return signals, nil
}

// findPullbackWindows 找出所有放量上涨事件对应的回调窗口
func (v *VolumeSurgePullback) findPullbackWindows(
	klines []*model.StockKline,
	volumes []int64,
	vma []float64,
) []pullbackWindow {
	cfg := v.Config
	n := len(klines)
	var windows []pullbackWindow

	for i := cfg.VolumeMAPeriod; i < n; i++ {
		if vma[i] == 0 {
			continue
		}
		volRatio := float64(volumes[i]) / vma[i]
		rallyPct := (klines[i].Close - klines[i].Open) / klines[i].Open * 100
		if volRatio < cfg.MinVolumeRatio || rallyPct < cfg.MinRallyPct {
			continue
		}

		// 找峰值
		surgeIdx := i
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

		days := peakIdx - surgeIdx
		windows = append(windows, pullbackWindow{
			surgeIdx:       surgeIdx,
			peakIdx:        peakIdx,
			peakPrice:      peakPrice,
			pullbackPct:    0,
			pullbackDays:   days,
			avgPullbackVol: 0,
		})

		// 跳过已覆盖的区间
		if peakIdx > i {
			i = peakIdx
		}
	}

	return windows
}

func (v *VolumeSurgePullback) calculateScore(
	klines []*model.StockKline,
	volumes []int64,
	vma []float64,
	w *pullbackWindow,
	currentIdx int,
) (float64, map[string]float64) {
	cfg := v.Config
	weights := cfg.Weights

	volRatio := float64(volumes[w.surgeIdx]) / vma[w.surgeIdx]
	volumeSurgeScore := math.Min(volRatio, 3.0) / 3.0 * 100

	rallyPct := (klines[w.peakIdx].Close - klines[w.surgeIdx].Open) / klines[w.surgeIdx].Open * 100
	priceRallyScore := math.Min(rallyPct, 8.0) / 8.0 * 100

	avgPullbackVol := v.avgVolume(volumes, w.peakIdx+1, currentIdx)
	pullbackVolRatio := 0.0
	if volumes[w.surgeIdx] > 0 {
		pullbackVolRatio = float64(avgPullbackVol) / float64(volumes[w.surgeIdx])
	}
	pullbackVolumeScore := math.Max(0, (1.0-pullbackVolRatio)*100)

	currentPrice := klines[currentIdx].Close
	peakPrice := klines[w.peakIdx].Close
	pullbackPct := (peakPrice - currentPrice) / peakPrice * 100
	pullbackDepthScore := math.Max(0, (1.0-pullbackPct/cfg.MaxPullbackPct)*100)

	total := volumeSurgeScore*weights.VolumeSurge +
		priceRallyScore*weights.PriceRally +
		pullbackVolumeScore*weights.PullbackVolume +
		pullbackDepthScore*weights.PullbackDepth

	subScores := map[string]float64{
		"volume_surge":    math.Round(volumeSurgeScore*100) / 100,
		"price_rally":     math.Round(priceRallyScore*100) / 100,
		"pullback_volume": math.Round(pullbackVolumeScore*100) / 100,
		"pullback_depth":  math.Round(pullbackDepthScore*100) / 100,
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
