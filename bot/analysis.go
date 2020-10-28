package bot

import (
	"cryptofu/bittrex"

	"github.com/shopspring/decimal"
)

func calculateSMA(tickers []bittrex.TickerResponse) (decimal.Decimal, error) {
	sma := decimal.NewFromInt(0)
	for i := 0; i < len(tickers); i++ {
		num, err := decimal.NewFromString(tickers[i].BidRate)
		if err != nil {
			return sma, err
		}
		sma = sma.Add(num).Div(decimal.NewFromInt(int64(len(tickers))))
	}
	return sma, nil
}

func calculateEMA() {

}

func tickerToTEMA(ticker bittrex.TickerResponse, lastVal decimal.Decimal) (decimal.Decimal, error) {

}
