package data

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"trading/model"
)

// FinancialReportRepo 定义财报数据访问接口
type FinancialReportRepo interface {
	Upsert(ctx context.Context, reports []*model.FinancialReport) error
	FindByCode(ctx context.Context, code string) ([]*model.FinancialReport, error)
	FindByCodeWithPagination(ctx context.Context, code string, limit, offset int) ([]*model.FinancialReport, error)
	FindAllCodes(ctx context.Context) ([]string, error)
}

type financialReportRepo struct {
	db *gorm.DB
}

func newFinancialReportRepo(db *gorm.DB) FinancialReportRepo {
	return &financialReportRepo{db: db}
}

// Upsert 批量插入或更新财报数据（code+report_date 联合唯一）
func (r *financialReportRepo) Upsert(ctx context.Context, reports []*model.FinancialReport) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}, {Name: "report_date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"report_type",
			"total_revenue", "total_cost", "net_profit", "net_profit_cut",
			"gross_margin", "net_margin", "operating_margin", "ebit_margin", "cost_profit_ratio",
			"roe", "roa",
			"asset_liability_ratio", "current_ratio", "quick_ratio",
			"total_asset_turnover", "inventory_turnover", "receivables_turnover",
			"operating_cash_flow", "operating_cash_flow_per_share",
			"eps", "bps",
			"updated_at",
		}),
	}).CreateInBatches(reports, 100).Error
}

// FindByCodeWithPagination 根据股票代码分页查询财报，按报告日期降序返回
func (r *financialReportRepo) FindByCodeWithPagination(ctx context.Context, code string, limit, offset int) ([]*model.FinancialReport, error) {
	var reports []*model.FinancialReport
	if err := r.db.WithContext(ctx).Where("code = ?", code).Order("report_date DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

// FindByCode 根据股票代码查询所有财报，按报告日期降序返回
func (r *financialReportRepo) FindByCode(ctx context.Context, code string) ([]*model.FinancialReport, error) {
	var reports []*model.FinancialReport
	if err := r.db.WithContext(ctx).Where("code = ?", code).Order("report_date DESC").Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

// FindAllCodes 查询所有 distinct 的股票代码
func (r *financialReportRepo) FindAllCodes(ctx context.Context) ([]string, error) {
	var codes []string
	if err := r.db.WithContext(ctx).Model(&model.FinancialReport{}).Distinct("code").Pluck("code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}
