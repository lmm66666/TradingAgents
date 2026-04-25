package business

import (
	"context"
	"fmt"

	"trading/data"
	"trading/model"
	"trading/pkg/broker"
)

const defaultFinancialReportYears = 5
const defaultFinancialReportNum = defaultFinancialReportYears * 4

type StockService interface {
	SaveHistoricalData(ctx context.Context, code string) error
	AppendStockData(ctx context.Context, code string) error
	SaveFinancialReportData(ctx context.Context, code string) error
	AppendFinancialReportData(ctx context.Context, code string) error
}

type stockService struct {
	broker       broker.IBroker
	dailyRepo    data.StockKlineDailyRepo
	weeklyRepo   data.StockKlineWeeklyRepo
	financialRepo data.FinancialReportRepo
}

// NewStockService 创建 StockService 实例
func NewStockService(b broker.IBroker, dailyRepo data.StockKlineDailyRepo, weeklyRepo data.StockKlineWeeklyRepo, financialRepo data.FinancialReportRepo) StockService {
	return &stockService{broker: b, dailyRepo: dailyRepo, weeklyRepo: weeklyRepo, financialRepo: financialRepo}
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

// AppendStockData 增量拉取并保存缺失的股票数据
// 对比数据库最新日期，只拉取并保存新增的数据，避免全量 upsert 导致的主键冲突
func (s *stockService) AppendStockData(ctx context.Context, code string) error {
	symbol, err := toSymbol(code)
	if err != nil {
		return err
	}

	if err := s.appendDaily(ctx, symbol, code); err != nil {
		return err
	}
	if err := s.appendWeekly(ctx, symbol, code); err != nil {
		return err
	}
	return nil
}

func (s *stockService) appendDaily(ctx context.Context, symbol, code string) error {
	var lastDate string
	if latest, err := s.dailyRepo.FindLatestByCode(ctx, code); err == nil {
		lastDate = latest.Date
	}

	klines, err := s.broker.GetStockHistorical(ctx, symbol, 240, 30)
	if err != nil {
		return fmt.Errorf("fetch daily failed: %w", err)
	}

	newKlines := filterAfterDate(cleanKlines(klines), lastDate)
	if len(newKlines) == 0 {
		return nil
	}

	if err := s.dailyRepo.Upsert(ctx, toDaily(newKlines)); err != nil {
		return fmt.Errorf("upsert daily failed: %w", err)
	}
	return nil
}

func (s *stockService) appendWeekly(ctx context.Context, symbol, code string) error {
	var lastDate string
	if latest, err := s.weeklyRepo.FindLatestByCode(ctx, code); err == nil {
		lastDate = latest.Date
	}

	klines, err := s.broker.GetStockHistorical(ctx, symbol, 1680, 10)
	if err != nil {
		return fmt.Errorf("fetch weekly failed: %w", err)
	}

	newKlines := filterAfterDate(filterIncompleteWeekly(cleanKlines(klines)), lastDate)
	if len(newKlines) == 0 {
		return nil
	}

	if err := s.weeklyRepo.Upsert(ctx, toWeekly(newKlines)); err != nil {
		return fmt.Errorf("upsert weekly failed: %w", err)
	}
	return nil
}


// AppendFinancialReportData 增量拉取并保存缺失的财报数据
// 对比数据库已有 report_date，只保存缺失的季度，避免全量 upsert
func (s *stockService) AppendFinancialReportData(ctx context.Context, code string) error {
	symbol, err := toSymbol(code)
	if err != nil {
		return err
	}

	existing, err := s.financialRepo.FindByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("find existing reports failed: %w", err)
	}

	existingDates := make(map[string]struct{}, len(existing))
	for _, r := range existing {
		existingDates[r.ReportDate] = struct{}{}
	}

	// 拉取最近 4 份（1年）
	reports, _, err := s.broker.GetFinancialReportHistorical(ctx, symbol, 1, 4)
	if err != nil {
		return fmt.Errorf("fetch financial report failed: %w", err)
	}

	var newReports []*model.FinancialReport
	for _, r := range reports {
		if _, ok := existingDates[r.ReportDate]; !ok {
			newReports = append(newReports, r)
		}
	}

	if len(newReports) == 0 {
		return nil
	}

	if err := s.financialRepo.Upsert(ctx, newReports); err != nil {
		return fmt.Errorf("upsert financial report failed: %w", err)
	}

	return nil
}

// SaveFinancialReportData 从 broker 获取5年财报数据并保存到 DB
// 一年4个季度，5年共20份财报
func (s *stockService) SaveFinancialReportData(ctx context.Context, code string) error {
	symbol, err := toSymbol(code)
	if err != nil {
		return err
	}

	reports, _, err := s.broker.GetFinancialReportHistorical(ctx, symbol, 1, defaultFinancialReportNum)
	if err != nil {
		return fmt.Errorf("fetch financial report failed: %w", err)
	}

	if len(reports) == 0 {
		return nil
	}

	if err := s.financialRepo.Upsert(ctx, reports); err != nil {
		return fmt.Errorf("upsert financial report failed: %w", err)
	}

	return nil
}
