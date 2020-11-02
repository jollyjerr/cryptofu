package bot

import "errors"

var (
	// ErrPing means the api poke failed
	ErrPing = errors.New("ping")
	// ErrCandles means the api call to get a market candles failed
	ErrCandles = errors.New("candles")
	// ErrTicker means the api call to get a market ticker failed
	ErrTicker = errors.New("ticker")
	// ErrCalcMACDNotEnoughInfo means that you tried to calculate a macd without enough information
	ErrCalcMACDNotEnoughInfo = errors.New("Not enough info to calculate a MACD value")
)
