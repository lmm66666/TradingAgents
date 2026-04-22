package data

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"trading/model"
)

// StockKlineWeeklyRepo 定义 StockKlineWeekly 数据访问接口
type StockKlineWeeklyRepo interface {
	Create(ctx context.Context, kline *model.StockKlineWeekly) error
	CreateBatch(ctx context.Context, klines []*model.StockKlineWeekly) error
	Upsert(ctx context.Context, klines []*model.StockKlineWeekly) error
	FindByID(ctx context.Context, id uint) (*model.StockKlineWeekly, error)
	FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKlineWeekly, error)
	Update(ctx context.Context, kline *model.StockKlineWeekly) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*model.StockKlineWeekly, error)
}

type stockKlineWeeklyRepo struct {
	db *gorm.DB
}

func newStockKlineWeeklyRepo(db *gorm.DB) StockKlineWeeklyRepo {
	return &stockKlineWeeklyRepo{db: db}
}

// Create 插入单条周线数据
func (r *stockKlineWeeklyRepo) Create(ctx context.Context, kline *model.StockKlineWeekly) error {
	return r.db.WithContext(ctx).Create(kline).Error
}

// CreateBatch 批量插入周线数据
func (r *stockKlineWeeklyRepo) CreateBatch(ctx context.Context, klines []*model.StockKlineWeekly) error {
	return r.db.WithContext(ctx).CreateInBatches(klines, 100).Error
}

// Upsert 批量插入或更新周线数据（code+date 联合唯一）
func (r *stockKlineWeeklyRepo) Upsert(ctx context.Context, klines []*model.StockKlineWeekly) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume", "updated_at"}),
	}).CreateInBatches(klines, 100).Error
}

// FindByID 根据主键查询周线
func (r *stockKlineWeeklyRepo) FindByID(ctx context.Context, id uint) (*model.StockKlineWeekly, error) {
	var kline model.StockKlineWeekly
	if err := r.db.WithContext(ctx).First(&kline, id).Error; err != nil {
		return nil, err
	}
	return &kline, nil
}

// FindByCode 根据股票代码查询周线，按日期升序
func (r *stockKlineWeeklyRepo) FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKlineWeekly, error) {
	var klines []*model.StockKlineWeekly
	query := r.db.WithContext(ctx).Where("code = ?", code).Order("date ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&klines).Error; err != nil {
		return nil, err
	}
	return klines, nil
}

// Update 更新周线数据
func (r *stockKlineWeeklyRepo) Update(ctx context.Context, kline *model.StockKlineWeekly) error {
	return r.db.WithContext(ctx).Save(kline).Error
}

// Delete 软删除指定周线记录
func (r *stockKlineWeeklyRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.StockKlineWeekly{}, id).Error
}

// List 分页查询所有周线数据
func (r *stockKlineWeeklyRepo) List(ctx context.Context, limit, offset int) ([]*model.StockKlineWeekly, error) {
	var klines []*model.StockKlineWeekly
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&klines).Error; err != nil {
		return nil, err
	}
	return klines, nil
}
