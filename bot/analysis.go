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

// CandlesToSMA calculates SMA from a slice of candles
func CandlesToSMA(candles []bittrex.CandleResponse) (decimal.Decimal, error) {
	sma := decimal.NewFromInt(0)
	for i := 0; i < len(candles); i++ {
		num, err := decimal.NewFromString(candles[i].Close)
		if err != nil {
			return sma, err
		}
		sma = sma.Add(num)
	}
	sma = sma.Div(decimal.NewFromInt(int64(len(candles))))
	return sma, nil
}

// DecimalsToSMA calculates SMA from a slice of decimal.Decimals
func DecimalsToSMA(nums []decimal.Decimal) decimal.Decimal {
	sma := decimal.NewFromInt(0)
	for i := 0; i < len(nums); i++ {
		sma = sma.Add(nums[i])
	}
	sma = sma.Div(decimal.NewFromInt(int64(len(nums))))
	return sma
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

// CandleToEMA converts a candle value to an EMA value
func CandleToEMA(candle bittrex.CandleResponse, lastVal decimal.Decimal, smoothing decimal.Decimal) (decimal.Decimal, error) {
	bid, err := decimal.NewFromString(candle.Close)
	if err != nil {
		return decimal.Zero, err
	}
	return CalculateEMA(bid, lastVal, smoothing), nil
}

// CandleToTEMA converts a candle value into a TEMA value
func CandleToTEMA(candle bittrex.CandleResponse, lastVal decimal.Decimal, smoothing decimal.Decimal) (decimal.Decimal, error) {
	bid, err := decimal.NewFromString(candle.Close)
	if err != nil {
		return decimal.Zero, err
	}
	return CalculateTEMA(bid, lastVal, smoothing), nil
}

// CalculateMACD calculates a macd value from a slice of tickers
func CalculateMACD(forThis decimal.Decimal, fromThese []bittrex.CandleResponse) (decimal.Decimal, error) {
	// check data
	if len(fromThese) < 26 {
		return decimal.Zero, ErrCalcMACDNotEnoughInfo
	}
	// 12 period ema
	sma1, err := CandlesToSMA(fromThese[len(fromThese)-12:])
	if err != nil {
		return decimal.Zero, err
	}
	ema12P := CalculateEMA(forThis, sma1, p12Smoothing)
	// 26 period ema
	sma2, err := CandlesToSMA(fromThese[len(fromThese)-26:])
	if err != nil {
		return decimal.Zero, err
	}
	ema26P := CalculateEMA(forThis, sma2, p26Smoothing)
	// return the result of the MACD formula with these values
	return ema12P.Sub(ema26P), nil
}

// CalculateSignalLine calculates a signal line
func CalculateSignalLine(fromThese []decimal.Decimal, lastVal decimal.Decimal) (decimal.Decimal, error) {
	signalPeriod := 9 // MACD Signal line uses 9 day EMA
	// Check data
	if len(fromThese) < signalPeriod {
		return decimal.Zero, ErrCalcSignalNotEnoughInfo
	} else if len(fromThese) == signalPeriod {
		sma := DecimalsToSMA(fromThese)
		return CalculateEMA(fromThese[len(fromThese)-1], sma, CalculateEMASmoothing(signalPeriod)), nil
	} else {
		return CalculateEMA(fromThese[len(fromThese)-1], lastVal, CalculateEMASmoothing(signalPeriod)), nil
	}
}

// CalculateHistogram is here so I don't forget that the equation is this simple
func CalculateHistogram(macd decimal.Decimal, signal decimal.Decimal) decimal.Decimal {
	return macd.Sub(signal)
}
