package bittrex

import "errors"

var (
	candleCache        = []CandleResponse{}
	lastYearRequested  = 2020
	lastMonthRequested = 6
	lastDayRequested   = 3
)

func getCandleResponse(requestNumber int, symbol string) ([]CandleResponse, error) {
	// CASE: first request while bot is setting up
	if requestNumber == 1 {
		result, err := GetHistoricalCandles(symbol, CandleIntervals["1min"], lastYearRequested, lastMonthRequested, 1)
		if err != nil {
			return []CandleResponse{}, err
		}
		future, err := GetHistoricalCandles(symbol, CandleIntervals["1min"], lastYearRequested, lastMonthRequested, 2)
		if err != nil {
			return []CandleResponse{}, err
		}
		candleCache = future
		return result, nil
	}

	// CASE: We already have the days data cached
	if len(candleCache) > requestNumber {
		data := candleCache[requestNumber]
		return []CandleResponse{data}, nil
	}

	// CASE: We need new data
	lastDayRequested++
	if lastDayRequested == 7 {
		return []CandleResponse{}, errors.New("boi") // temporary solution for stopping tests
	}
	future, err := GetHistoricalCandles(symbol, CandleIntervals["1min"], lastYearRequested, lastMonthRequested, lastDayRequested)
	if err != nil {
		return []CandleResponse{}, err
	}
	candleCache = append(candleCache, future...)
	data := candleCache[requestNumber]
	return []CandleResponse{data}, nil
	// return []CandleResponse{}, errors.New("boi")
}
