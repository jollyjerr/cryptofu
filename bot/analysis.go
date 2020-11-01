package bot

import (
	"cryptofu/bittrex"

	"github.com/shopspring/decimal"
)

var (
	one   = decimal.NewFromInt(1)
	two   = decimal.NewFromInt(2)
	three = decimal.NewFromInt(3)

	p12Smoothing = CalculateEMASmoothing(12)
	p26Smoothing = CalculateEMASmoothing(26)
)

// CalculateSMA calculates SMA from a slice of tickers
func CalculateSMA(tickers []bittrex.TickerResponse) (decimal.Decimal, error) {
	sma := decimal.NewFromInt(0)
	for i := 0; i < len(tickers); i++ {
		num, err := decimal.NewFromString(tickers[i].BidRate)
		if err != nil {
			return sma, err
		}
		sma = sma.Add(num)
	}
	sma = sma.Div(decimal.NewFromInt(int64(len(tickers))))
	return sma, nil
}

// CalculateEMASmoothing calculates EMA smoothing
func CalculateEMASmoothing(period int) decimal.Decimal {
	return two.Div(decimal.NewFromInt(int64(period + 1)))
}

// CalculateEMA calculates an EMA
func CalculateEMA(forThis decimal.Decimal, basedOn decimal.Decimal, smoothing decimal.Decimal) decimal.Decimal {
	return forThis.Mul(smoothing).Add(basedOn.Mul(one.Sub(smoothing)))
}

// CalculateTEMA calculates a TEMA
func CalculateTEMA(forThis decimal.Decimal, basedOn decimal.Decimal, smoothing decimal.Decimal) decimal.Decimal {
	EMA1 := CalculateEMA(forThis, basedOn, smoothing)
	EMA2 := CalculateEMA(EMA1, basedOn, smoothing)
	EMA3 := CalculateEMA(EMA2, basedOn, smoothing)
	return three.Mul(EMA1).Sub(three.Mul(EMA2)).Add(EMA3)
}

// TickerToEMA converts a ticker value to an EMA value
func TickerToEMA(ticker bittrex.TickerResponse, lastVal decimal.Decimal, smoothing decimal.Decimal) (decimal.Decimal, error) {
	bid, err := decimal.NewFromString(ticker.BidRate)
	if err != nil {
		return decimal.Zero, err
	}
	return CalculateEMA(bid, lastVal, smoothing), nil
}

// TickerToTEMA converts a ticker value into a TEMA value
func TickerToTEMA(ticker bittrex.TickerResponse, lastVal decimal.Decimal, smoothing decimal.Decimal) (decimal.Decimal, error) {
	bid, err := decimal.NewFromString(ticker.BidRate)
	if err != nil {
		return decimal.Zero, err
	}
	return CalculateTEMA(bid, lastVal, smoothing), nil
}

// CalculateMACD calculates a macd value from a slice of tickers
func CalculateMACD(forThis decimal.Decimal, fromThese []bittrex.TickerResponse) (decimal.Decimal, error) {
	// check data
	if len(fromThese) < 26 {
		return decimal.Zero, ErrCalcMACDNotEnoughInfo
	}
	// 12 period ema
	sma1, err := CalculateSMA(fromThese[13:])
	if err != nil {
		return decimal.Zero, err
	}
	ema12P := CalculateEMA(forThis, sma1, p12Smoothing)
	// 26 period ema
	sma2, err := CalculateSMA(fromThese[27:])
	if err != nil {
		return decimal.Zero, err
	}
	ema26P := CalculateEMA(forThis, sma2, p26Smoothing)
	// return the result of the MACD formula with these values
	return ema12P.Sub(ema26P), nil
}
