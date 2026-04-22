package data

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"trading/model"
)

// StockKlineDailyRepo 定义 StockKlineDaily 数据访问接口
type StockKlineDailyRepo interface {
	Create(ctx context.Context, kline *model.StockKlineDaily) error
	CreateBatch(ctx context.Context, klines []*model.StockKlineDaily) error
	Upsert(ctx context.Context, klines []*model.StockKlineDaily) error
	FindByID(ctx context.Context, id uint) (*model.StockKlineDaily, error)
	FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKlineDaily, error)
	FindLatestByCode(ctx context.Context, code string) (*model.StockKlineDaily, error)
	FindAllCodes(ctx context.Context) ([]string, error)
	Update(ctx context.Context, kline *model.StockKlineDaily) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*model.StockKlineDaily, error)
}

type stockKlineDailyRepo struct {
	db *gorm.DB
}

func newStockKlineDailyRepo(db *gorm.DB) StockKlineDailyRepo {
	return &stockKlineDailyRepo{db: db}
}

// Create 插入单条日线数据
func (r *stockKlineDailyRepo) Create(ctx context.Context, kline *model.StockKlineDaily) error {
	return r.db.WithContext(ctx).Create(kline).Error
}

// CreateBatch 批量插入日线数据
func (r *stockKlineDailyRepo) CreateBatch(ctx context.Context, klines []*model.StockKlineDaily) error {
	return r.db.WithContext(ctx).CreateInBatches(klines, 100).Error
}

// Upsert 批量插入或更新日线数据（code+date 联合唯一）
func (r *stockKlineDailyRepo) Upsert(ctx context.Context, klines []*model.StockKlineDaily) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume", "updated_at"}),
	}).CreateInBatches(klines, 100).Error
}

// FindByID 根据主键查询日线
func (r *stockKlineDailyRepo) FindByID(ctx context.Context, id uint) (*model.StockKlineDaily, error) {
	var kline model.StockKlineDaily
	if err := r.db.WithContext(ctx).First(&kline, id).Error; err != nil {
		return nil, err
	}
	return &kline, nil
}

// FindByCode 根据股票代码查询最近 limit 条日线，按日期升序返回
func (r *stockKlineDailyRepo) FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKlineDaily, error) {
	var klines []*model.StockKlineDaily
	query := r.db.WithContext(ctx).Where("code = ?", code).Order("date DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&klines).Error; err != nil {
		return nil, err
	}
	for i, j := 0, len(klines)-1; i < j; i, j = i+1, j-1 {
		klines[i], klines[j] = klines[j], klines[i]
	}
	return klines, nil
}

// Update 更新日线数据
func (r *stockKlineDailyRepo) Update(ctx context.Context, kline *model.StockKlineDaily) error {
	return r.db.WithContext(ctx).Save(kline).Error
}

// Delete 软删除指定日线记录
func (r *stockKlineDailyRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.StockKlineDaily{}, id).Error
}

// FindLatestByCode 查询指定代码最新一条日线（按日期降序）
func (r *stockKlineDailyRepo) FindLatestByCode(ctx context.Context, code string) (*model.StockKlineDaily, error) {
	var kline model.StockKlineDaily
	if err := r.db.WithContext(ctx).Where("code = ?", code).Order("date DESC").First(&kline).Error; err != nil {
		return nil, err
	}
	return &kline, nil
}

// FindAllCodes 查询所有 distinct 的股票代码
func (r *stockKlineDailyRepo) FindAllCodes(ctx context.Context) ([]string, error) {
	var codes []string
	if err := r.db.WithContext(ctx).Model(&model.StockKlineDaily{}).Distinct("code").Pluck("code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

// List 分页查询所有日线数据
func (r *stockKlineDailyRepo) List(ctx context.Context, limit, offset int) ([]*model.StockKlineDaily, error) {
	var klines []*model.StockKlineDaily
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&klines).Error; err != nil {
		return nil, err
	}
	return klines, nil
}
