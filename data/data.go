package data

import (
	"fmt"

	"trading/config"
	"trading/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Data struct {
	db *gorm.DB
}

// New 创建 Data 实例，内部根据配置初始化 gorm.DB 连接
func New(cfg config.DB) (*Data, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql failed: %w", err)
	}

	if err := db.AutoMigrate(&model.StockKline{}); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	return &Data{db: db}, nil
}

// DB 返回底层 gorm.DB 实例
func (d *Data) DB() *gorm.DB {
	return d.db
}

// StockKline 返回 StockKline 的 Repository
func (d *Data) StockKline() StockKlineRepo {
	return newStockKlineRepo(d.db)
}
