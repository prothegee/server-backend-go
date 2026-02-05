/*
note:
- this implementation is based on dgkafka behaviour
  - https://github.com/prothegee/drogon-examples/tree/main/drogon-kafka
*/
package pkg

import (
	"math/rand"
)

const (
	GOKAFKA_STOCK_TRADE_TOPIC = "consume-stock-trade"
	GOKAFKA_DELAY_MS          = 1000
)

type StockTrade struct{}

type StockTrade_tj struct {
	Stock       float64 `json:"stock"`
	Currency    string  `json:"currency"`
	LastUpdated string  `json:"last_updated"`
	ID          string  `json:"id"`
}

func (_ StockTrade) StockTradeNew(stock float64, currency, id string) *StockTrade_tj {
	return &StockTrade_tj{
		Stock:       stock,
		Currency:    currency,
		LastUpdated: TimestampNow(),
		ID:          id,
	}
}

func (s *StockTrade_tj) Update() {
	r := rand.Intn(11)

	// only apply 1 (purchased) & 2 (sold) by entity
	switch r {
	case 1:
		{
			s.Stock += rand.Float64() * 3_000
		}
	case 2:
		{
			if s.Stock > 0 {
				s.Stock -= rand.Float64() * 3_000

				if s.Stock < 0 {
					s.Stock = 0
				}
			}
		}
	default:
		{
			// nothing todo
		}
	}
}
