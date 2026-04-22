package data

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"trading/model"
)

// StockKlineRepo 定义 StockKline 数据访问接口
type StockKlineRepo interface {
	Create(ctx context.Context, kline *model.StockKline) error
	CreateBatch(ctx context.Context, klines []*model.StockKline) error
	Upsert(ctx context.Context, klines []*model.StockKline) error
	FindByID(ctx context.Context, id uint) (*model.StockKline, error)
	FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKline, error)
	Update(ctx context.Context, kline *model.StockKline) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*model.StockKline, error)
}

type stockKlineRepo struct {
	db *gorm.DB
}

func newStockKlineRepo(db *gorm.DB) StockKlineRepo {
	return &stockKlineRepo{db: db}
}

// Create 插入单条 K 线数据
func (r *stockKlineRepo) Create(ctx context.Context, kline *model.StockKline) error {
	return r.db.WithContext(ctx).Create(kline).Error
}

// CreateBatch 批量插入 K 线数据
func (r *stockKlineRepo) CreateBatch(ctx context.Context, klines []*model.StockKline) error {
	return r.db.WithContext(ctx).CreateInBatches(klines, 100).Error
}

// Upsert 批量插入或更新 K 线数据（code+date 联合唯一）
func (r *stockKlineRepo) Upsert(ctx context.Context, klines []*model.StockKline) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume", "updated_at"}),
	}).CreateInBatches(klines, 100).Error
}

// FindByID 根据主键查询
func (r *stockKlineRepo) FindByID(ctx context.Context, id uint) (*model.StockKline, error) {
	var kline model.StockKline
	if err := r.db.WithContext(ctx).First(&kline, id).Error; err != nil {
		return nil, err
	}
	return &kline, nil
}

// FindByCode 根据股票代码查询，按日期升序
func (r *stockKlineRepo) FindByCode(ctx context.Context, code string, limit int) ([]*model.StockKline, error) {
	var klines []*model.StockKline
	query := r.db.WithContext(ctx).Where("code = ?", code).Order("date ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&klines).Error; err != nil {
		return nil, err
	}
	return klines, nil
}

// Update 更新 K 线数据
func (r *stockKlineRepo) Update(ctx context.Context, kline *model.StockKline) error {
	return r.db.WithContext(ctx).Save(kline).Error
}

// Delete 软删除指定记录
func (r *stockKlineRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.StockKline{}, id).Error
}

// List 分页查询所有 K 线数据
func (r *stockKlineRepo) List(ctx context.Context, limit, offset int) ([]*model.StockKline, error) {
	var klines []*model.StockKline
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&klines).Error; err != nil {
		return nil, err
	}
	return klines, nil
}
