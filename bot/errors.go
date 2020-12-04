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
	// ErrCalcSignalNotEnoughInfo means that you tried to calculate a signal without enough information
	ErrCalcSignalNotEnoughInfo = errors.New("Not enough info to calculate a signal line value")
	// ErrNetNewOrder means there was a network error while creating a new order
	ErrNetNewOrder = errors.New("Network error while creating a new order")
)
