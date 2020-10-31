package bot

import (
	"cryptofu/bittrex"

	"github.com/shopspring/decimal"
)

var (
	one   = decimal.NewFromInt(1)
	two   = decimal.NewFromInt(2)
	three = decimal.NewFromInt(3)
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

func calculateEMASmoothing(period int) decimal.Decimal {
	return two.Div(decimal.NewFromInt(int64(period + 1)))
}

func calculateEMA(forThis decimal.Decimal, basedOn decimal.Decimal, smoothing decimal.Decimal) decimal.Decimal {
	return forThis.Mul(smoothing).Add(basedOn.Mul(one.Sub(smoothing)))
}

func tickerToTEMA(ticker bittrex.TickerResponse, lastVal decimal.Decimal, smoothing decimal.Decimal) (decimal.Decimal, error) {
	bid, err := decimal.NewFromString(ticker.BidRate)
	if err != nil {
		return decimal.Zero, err
	}

	EMA1 := calculateEMA(bid, lastVal, smoothing)
	EMA2 := calculateEMA(EMA1, lastVal, smoothing)
	EMA3 := calculateEMA(EMA2, lastVal, smoothing)

	return three.Mul(EMA1).Sub(three.Mul(EMA2)).Add(EMA3), nil
}
