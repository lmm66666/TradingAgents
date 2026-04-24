package indicator_filter

import (
	"trading/model"
)

type IFilter interface {
	Filter(klines []*model.StockKline) []string
}
