package strategy

import (
	"trading/model"
	"trading/pkg/score"
)

type rule struct {
	rate  float64
	score score.IScorer
}

type Strategy struct {
	Name  string
	Rules []rule
}

func (s *Strategy) Scan(klines []*model.StockKline) ([]Signal, error) {

}
