package business

import (
	"context"
	"fmt"
	"strings"
	"time"

	"trading/data"
	"trading/model"
	"trading/pkg/broker"
	"trading/pkg/utils"
)

// MACDPoint MACD 指标点
type MACDPoint struct {
	DIF float64 `json:"dif"`
	DEA float64 `json:"dea"`
	BAR float64 `json:"bar"`
}

// KDJPoint KDJ 指标点
type KDJPoint struct {
	K float64 `json:"k"`
	D float64 `json:"d"`
	J float64 `json:"j"`
}

// AnalysisItem 按时间点聚合的分析数据（K线 + MACD + KDJ）
type AnalysisItem struct {
	Date   string    `json:"date"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
	KDJ    KDJPoint  `json:"kdj"`
	MACD   MACDPoint `json:"macd"`
}

type StockAnalysisData struct {
	Daily  []AnalysisItem `json:"daily"`
	Weekly []AnalysisItem `json:"weekly"`
}

type StockService interface {
	SaveHistoricalData(ctx context.Context, code string) error
	GetStockAnalysisData(ctx context.Context, code string) (*StockAnalysisData, error)
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

// GetStockAnalysisData 获取股票分析数据（日线/周线 + MACD/KDJ）
func (s *stockService) GetStockAnalysisData(ctx context.Context, code string) (*StockAnalysisData, error) {
	dailies, err := s.dailyRepo.FindByCode(ctx, code, 100)
	if err != nil {
		return nil, fmt.Errorf("find daily by code failed: %w", err)
	}

	weeklies, err := s.weeklyRepo.FindByCode(ctx, code, 50)
	if err != nil {
		return nil, fmt.Errorf("find weekly by code failed: %w", err)
	}

	dailyKlines := dailyToKlines(dailies)
	weeklyKlines := weeklyToKlines(weeklies)

	dailyMACD := utils.ComputeMACD(dailyKlines)
	weeklyMACD := utils.ComputeMACD(weeklyKlines)
	dailyKDJ := utils.ComputeKDJ(dailyKlines)
	weeklyKDJ := utils.ComputeKDJ(weeklyKlines)

	return &StockAnalysisData{
		Daily:  buildAnalysisItems(dailyKlines, dailyMACD, dailyKDJ),
		Weekly: buildAnalysisItems(weeklyKlines, weeklyMACD, weeklyKDJ),
	}, nil
}

func buildAnalysisItems(klines []*model.StockKline, macd []utils.MACDResult, kdj []utils.KDJResult) []AnalysisItem {
	result := make([]AnalysisItem, len(klines))
	for i, k := range klines {
		result[i] = AnalysisItem{
			Date:   k.Date,
			Open:   k.Open,
			High:   k.High,
			Low:    k.Low,
			Close:  k.Close,
			Volume: k.Volume,
			KDJ:    KDJPoint{K: kdj[i].K, D: kdj[i].D, J: kdj[i].J},
			MACD:   MACDPoint{DIF: macd[i].DIF, DEA: macd[i].DEA, BAR: macd[i].BAR},
		}
	}
	return result
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

func toDaily(klines []*model.StockKline) []*model.StockKlineDaily {
	result := make([]*model.StockKlineDaily, 0, len(klines))
	for _, k := range klines {
		d := model.StockKlineDaily(*k)
		result = append(result, &d)
	}
	return result
}

func toWeekly(klines []*model.StockKline) []*model.StockKlineWeekly {
	result := make([]*model.StockKlineWeekly, 0, len(klines))
	for _, k := range klines {
		w := model.StockKlineWeekly(*k)
		result = append(result, &w)
	}
	return result
}

func dailyToKlines(dailies []*model.StockKlineDaily) []*model.StockKline {
	result := make([]*model.StockKline, 0, len(dailies))
	for _, d := range dailies {
		k := model.StockKline(*d)
		result = append(result, &k)
	}
	return result
}

func weeklyToKlines(weeklies []*model.StockKlineWeekly) []*model.StockKline {
	result := make([]*model.StockKline, 0, len(weeklies))
	for _, w := range weeklies {
		k := model.StockKline(*w)
		result = append(result, &k)
	}
	return result
}
