package business

import (
	"context"
	"fmt"

	"trading/data"
	"trading/model"
)

// QueryService 股票数据查询服务
type QueryService interface {
	// FindStockPricesByCode 根据股票代码和周期查询 K 线数据
	FindStockPricesByCode(ctx context.Context, code, cycle string, limit, offset int) ([]*model.StockKlineDaily, error)
	// FindFinancialReportsByCode 根据股票代码分页查询财报数据
	FindFinancialReportsByCode(ctx context.Context, code string, limit, offset int) ([]*model.FinancialReport, error)
}

type queryService struct {
	dailyRepo     data.StockKlineDailyRepo
	weeklyRepo    data.StockKlineWeeklyRepo
	financialRepo data.FinancialReportRepo
}

// NewQueryService 创建 QueryService 实例
func NewQueryService(dailyRepo data.StockKlineDailyRepo, weeklyRepo data.StockKlineWeeklyRepo, financialRepo data.FinancialReportRepo) QueryService {
	return &queryService{dailyRepo: dailyRepo, weeklyRepo: weeklyRepo, financialRepo: financialRepo}
}

func (s *queryService) FindStockPricesByCode(ctx context.Context, code, cycle string, limit, offset int) ([]*model.StockKlineDaily, error) {
	switch cycle {
	case "daily":
		return s.dailyRepo.FindByCodeWithPagination(ctx, code, limit, offset)
	case "weekly":
		weeklies, err := s.weeklyRepo.FindByCodeWithPagination(ctx, code, limit, offset)
		if err != nil {
			return nil, err
		}
		result := make([]*model.StockKlineDaily, 0, len(weeklies))
		for _, w := range weeklies {
			d := model.StockKlineDaily(*w)
			result = append(result, &d)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported cycle: %s, expected daily or weekly", cycle)
	}
}

func (s *queryService) FindFinancialReportsByCode(ctx context.Context, code string, limit, offset int) ([]*model.FinancialReport, error) {
	return s.financialRepo.FindByCodeWithPagination(ctx, code, limit, offset)
}
