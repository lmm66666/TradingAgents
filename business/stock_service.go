package business

import (
	"context"
	"fmt"
	"strings"
	"time"

	"trading/data"
	"trading/model"
	"trading/pkg/broker"
)

// StockService 股票数据业务逻辑接口
type StockService interface {
	SaveHistoricalData(ctx context.Context, code string) error
	GetStockData(ctx context.Context, code string, scale int, length int) ([]*model.StockKline, error)
}

type stockService struct {
	broker     broker.IBroker
	dailyRepo  data.StockKlineDailyRepo
	weeklyRepo data.StockKlineWeeklyRepo
}

// NewStockService 创建 StockService 实例
func NewStockService(b broker.IBroker, dailyRepo data.StockKlineDailyRepo, weeklyRepo data.StockKlineWeeklyRepo) StockService {
	return &stockService{broker: b, dailyRepo: dailyRepo, weeklyRepo: weeklyRepo}
}

// SaveHistoricalData 从 broker 获取历史数据并保存到 DB
// 日线保存 1000 个点（scale=240），周线保存 200 个点（scale=1680）
// 周线最后一条若不是周五则丢弃，避免保存不完整周数据
func (s *stockService) SaveHistoricalData(ctx context.Context, code string) error {
	symbol, err := toSymbol(code)
	if err != nil {
		return err
	}

	// 拉取日线
	dailyKlines, err := s.broker.GetStockHistorical(ctx, symbol, 240, 1000)
	if err != nil {
		return fmt.Errorf("fetch daily historical failed: %w", err)
	}

	cleanedDaily := cleanKlines(dailyKlines)
	daily := toDaily(cleanedDaily)
	if err := s.dailyRepo.Upsert(ctx, daily); err != nil {
		return fmt.Errorf("upsert daily failed: %w", err)
	}

	// 拉取周线
	weeklyKlines, err := s.broker.GetStockHistorical(ctx, symbol, 1680, 200)
	if err != nil {
		return fmt.Errorf("fetch weekly historical failed: %w", err)
	}

	cleanedWeekly := cleanKlines(weeklyKlines)
	filteredWeekly := filterIncompleteWeekly(cleanedWeekly)
	weekly := toWeekly(filteredWeekly)
	if err := s.weeklyRepo.Upsert(ctx, weekly); err != nil {
		return fmt.Errorf("upsert weekly failed: %w", err)
	}

	return nil
}

// GetStockData 从 DB 获取股票数据并按 scale 聚合
func (s *stockService) GetStockData(ctx context.Context, code string, scale int, length int) ([]*model.StockKline, error) {
	if length <= 0 {
		length = 240
	}

	dailies, err := s.dailyRepo.FindByCode(ctx, code, 0)
	if err != nil {
		return nil, fmt.Errorf("find daily by code failed: %w", err)
	}

	klines := dailyToKlines(dailies)

	groupSize := max(1, scale/240)
	if groupSize == 1 {
		return takeLast(klines, length), nil
	}

	aggregated := aggregateKlines(klines, groupSize)
	return takeLast(aggregated, length), nil
}

// toSymbol 将纯数字 code 转换为带前缀的 symbol
func toSymbol(code string) (string, error) {
	switch {
	case strings.HasPrefix(code, "6"):
		return "sh" + code, nil
	case strings.HasPrefix(code, "0"), strings.HasPrefix(code, "3"):
		return "sz" + code, nil
	default:
		return "", fmt.Errorf("unsupported stock code: %s", code)
	}
}

// cleanKlines 去掉 Code 中的 sh/sz 前缀
func cleanKlines(klines []model.StockKline) []*model.StockKline {
	result := make([]*model.StockKline, 0, len(klines))
	for i := range klines {
		k := &klines[i]
		k.Code = strings.TrimPrefix(k.Code, "sh")
		k.Code = strings.TrimPrefix(k.Code, "sz")
		result = append(result, k)
	}
	return result
}

// filterIncompleteWeekly 过滤掉最后一条非周五的周线数据
// broker 返回的周线若最后一条是周中日期，则为不完整周，应丢弃
func filterIncompleteWeekly(klines []*model.StockKline) []*model.StockKline {
	if len(klines) == 0 {
		return klines
	}

	last := klines[len(klines)-1]
	if !isFriday(last.Date) {
		return klines[:len(klines)-1]
	}
	return klines
}

// isFriday 判断日期字符串是否为周五
// 支持格式：2006-01-02 或 2006-01-02 15:04:05
func isFriday(dateStr string) bool {
	layout := "2006-01-02"
	if len(dateStr) > 10 {
		layout = "2006-01-02 15:04:05"
	}
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		return false
	}
	return t.Weekday() == time.Friday
}

// toDaily 将通用 K 线转换为日线模型
func toDaily(klines []*model.StockKline) []*model.StockKlineDaily {
	result := make([]*model.StockKlineDaily, 0, len(klines))
	for _, k := range klines {
		result = append(result, &model.StockKlineDaily{
			Code:   k.Code,
			Date:   k.Date,
			Open:   k.Open,
			High:   k.High,
			Low:    k.Low,
			Close:  k.Close,
			Volume: k.Volume,
		})
	}
	return result
}

// toWeekly 将通用 K 线转换为周线模型
func toWeekly(klines []*model.StockKline) []*model.StockKlineWeekly {
	result := make([]*model.StockKlineWeekly, 0, len(klines))
	for _, k := range klines {
		result = append(result, &model.StockKlineWeekly{
			Code:   k.Code,
			Date:   k.Date,
			Open:   k.Open,
			High:   k.High,
			Low:    k.Low,
			Close:  k.Close,
			Volume: k.Volume,
		})
	}
	return result
}

// dailyToKlines 将日线模型转换为通用 K 线
func dailyToKlines(dailies []*model.StockKlineDaily) []*model.StockKline {
	result := make([]*model.StockKline, 0, len(dailies))
	for _, d := range dailies {
		result = append(result, &model.StockKline{
			Code:   d.Code,
			Date:   d.Date,
			Open:   d.Open,
			High:   d.High,
			Low:    d.Low,
			Close:  d.Close,
			Volume: d.Volume,
		})
	}
	return result
}

// aggregateKlines 将 klines 按 groupSize 条聚合成一条
func aggregateKlines(klines []*model.StockKline, groupSize int) []*model.StockKline {
	if len(klines) == 0 || groupSize <= 1 {
		return klines
	}

	result := make([]*model.StockKline, 0, (len(klines)+groupSize-1)/groupSize)
	for i := 0; i < len(klines); i += groupSize {
		end := i + groupSize
		if end > len(klines) {
			end = len(klines)
		}

		agg := aggregateGroup(klines[i:end])
		result = append(result, agg)
	}
	return result
}

// aggregateGroup 聚合一组 K 线数据
func aggregateGroup(group []*model.StockKline) *model.StockKline {
	first := group[0]
	last := group[len(group)-1]

	var high, low float64
	var volume int64
	for _, k := range group {
		if k.High > high || high == 0 {
			high = k.High
		}
		if k.Low < low || low == 0 {
			low = k.Low
		}
		volume += k.Volume
	}

	return &model.StockKline{
		Code:   first.Code,
		Date:   last.Date,
		Open:   first.Open,
		High:   high,
		Low:    low,
		Close:  last.Close,
		Volume: volume,
	}
}

// takeLast 取切片最后 n 条
func takeLast(klines []*model.StockKline, n int) []*model.StockKline {
	if n >= len(klines) {
		return klines
	}
	return klines[len(klines)-n:]
}
