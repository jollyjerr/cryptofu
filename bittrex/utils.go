package bittrex

var (
	candleCache        = []CandleResponse{}
	lastYearRequested  = 2020
	lastMonthRequested = 1
	lastDayRequested   = 2
)

func getCandleResponse(requestNumber int) ([]CandleResponse, error) {
	// CASE: first request while bot is setting up
	if requestNumber == 1 {
		result, err := GetHistoricalCandles(Symbols["Bitcoin"], CandleIntervals["1min"], 2020, 1, 1)
		if err != nil {
			return []CandleResponse{}, err
		}
		future, err := GetHistoricalCandles(Symbols["Bitcoin"], CandleIntervals["1min"], lastYearRequested, lastMonthRequested, lastDayRequested)
		if err != nil {
			return []CandleResponse{}, err
		}
		candleCache = future
		lastDayRequested = 2
		lastMonthRequested = 1
		lastYearRequested = 2020
		return result, nil
	}

	// CASE: We already have the days data cached
	if len(candleCache) > requestNumber {
		data := candleCache[requestNumber]
		return []CandleResponse{data}, nil
	}

	// CASE: We need new data
	lastDayRequested++
	future, err := GetHistoricalCandles(Symbols["Bitcoin"], CandleIntervals["1min"], lastYearRequested, lastMonthRequested, lastDayRequested)
	if err != nil {
		return []CandleResponse{}, err
	}
	candleCache = append(candleCache, future...)
	data := candleCache[requestNumber]
	return []CandleResponse{data}, nil
}
