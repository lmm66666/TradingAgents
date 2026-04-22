package data

import "gorm.io/gorm"

// Data 持有数据库连接，提供各模型 Repository 的入口
type Data struct {
	db *gorm.DB
}

// New 创建 Data 实例
func New(db *gorm.DB) *Data {
	return &Data{db: db}
}

// DB 返回底层 gorm.DB 实例
func (d *Data) DB() *gorm.DB {
	return d.db
}

// StockKline 返回 StockKline 的 Repository
func (d *Data) StockKline() StockKlineRepo {
	return newStockKlineRepo(d.db)
}
