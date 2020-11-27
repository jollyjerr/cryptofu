package bot

import (
	"cryptofu/bittrex"
	"fmt"

	"github.com/shopspring/decimal"
)

var (
	buyHistory  = make([]bittrex.CandleResponse, 0)
	sellHistory = make([]bittrex.CandleResponse, 0)
)

func printStats() {
	for i := 0; i < len(sellHistory); i++ {
		buy, err := decimal.NewFromString(buyHistory[i].Close)
		if err != nil {
			continue
		}
		sell, err := decimal.NewFromString(sellHistory[i].Close)
		if err != nil {
			continue
		}
		diff := sell.Sub(buy)
		fmt.Println("💰", diff.StringFixed(2))
	}
}

func saveBuy(buy bittrex.CandleResponse) {
	buyHistory = append(buyHistory, buy)
}

func saveSell(sell bittrex.CandleResponse) {
	sellHistory = append(sellHistory, sell)
}
