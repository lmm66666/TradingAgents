package business

import (
	"context"
	"fmt"
	"strings"

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
	broker broker.IBroker
	repo   data.StockKlineRepo
}

// NewStockService 创建 StockService 实例
func NewStockService(b broker.IBroker, r data.StockKlineRepo) StockService {
	return &stockService{broker: b, repo: r}
}

// SaveHistoricalData 从 broker 获取历史数据并保存到 DB
func (s *stockService) SaveHistoricalData(ctx context.Context, code string) error {
	symbol, err := toSymbol(code)
	if err != nil {
		return err
	}

	klines, err := s.broker.GetStockHistorical(ctx, symbol, 240, 300)
	if err != nil {
		return fmt.Errorf("fetch historical failed: %w", err)
	}

	cleaned := cleanKlines(klines)
	if err := s.repo.Upsert(ctx, cleaned); err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}

	return nil
}

// GetStockData 从 DB 获取股票数据并按 scale 聚合
func (s *stockService) GetStockData(ctx context.Context, code string, scale int, length int) ([]*model.StockKline, error) {
	if length <= 0 {
		length = 240
	}

	klines, err := s.repo.FindByCode(ctx, code, 0)
	if err != nil {
		return nil, fmt.Errorf("find by code failed: %w", err)
	}

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
