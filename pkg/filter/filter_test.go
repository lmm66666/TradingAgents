package filter

import (
	"fmt"

	"trading/model"
)

func buildKlines(n int) []*model.StockKline {
	klines := make([]*model.StockKline, n)
	for i := 0; i < n; i++ {
		price := 10.0 + float64(i)*0.01
		klines[i] = &model.StockKline{
			Code:   "600312",
			Date:   fmt.Sprintf("2026-%02d-%02d", (i/30)+1, (i%30)+1),
			Open:   price - 0.05,
			High:   price + 0.1,
			Low:    price - 0.1,
			Close:  price,
			Volume: 100000,
		}
	}
	return klines
}
