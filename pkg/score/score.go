package score

import (
	"trading/model"
)

type Result struct {
	Date  string
	Score int // 范围 0-100
}

type IScorer interface {
	Name() string
	Score(klines []*model.StockKline) []*Result
}
