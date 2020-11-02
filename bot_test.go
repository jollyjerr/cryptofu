package main

import (
	"cryptofu/bittrex"
	"cryptofu/bot"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

var (
	exampleCandles = func() []bittrex.CandleResponse {
		data := make([]bittrex.CandleResponse, 0)
		for i := 1; i < 31; i++ {
			data = append(data, createDemoCandleResponse(i))
		}
		return data
	}()
)

func createDemoCandleResponse(i int) bittrex.CandleResponse {
	return bittrex.CandleResponse{
		Close: fmt.Sprintf("%d", 20000+i),
	}
}

func checkStringFixed(forThis decimal.Decimal, fixedAmt int32, expected string, t *testing.T) {
	if forThis.StringFixed(fixedAmt) != expected {
		t.Errorf("Was %s, expected %s", forThis.StringFixed(fixedAmt), expected)
	}
}

func td(num int64) decimal.Decimal {
	return decimal.NewFromInt(num)
}

/*
	analysis.go
*/

func TestCandlesToSMA(t *testing.T) {
	got, err := bot.CandlesToSMA(exampleCandles)
	if err != nil {
		t.Error(err)
	}
	checkStringFixed(got, 2, "20015.50", t)
}

func TestDecimalsToSMA(t *testing.T) {
	// TODO
}

func TestCalculateEMASmoothing(t *testing.T) {
	got := bot.CalculateEMASmoothing(6)
	checkStringFixed(got, 2, "0.29", t)
}

func TestCalculateEMA(t *testing.T) {
	number := decimal.NewFromInt(20005)
	lastVal := decimal.NewFromInt(20000)
	got := bot.CalculateEMA(number, lastVal, bot.CalculateEMASmoothing(2))
	checkStringFixed(got, 2, "20003.33", t)
}

func TestCalculateTEMA(t *testing.T) {
	// TODO
}

func TestCandleToEMA(t *testing.T) {
	// TODO
}

func TestCandleToTEMA(t *testing.T) {
	// TODO
}

func TestCalculateMACD(t *testing.T) {
	// Positive
	number := decimal.NewFromInt(20049)
	got, err := bot.CalculateMACD(number, exampleCandles)
	if err != nil {
		t.Error(err)
	}
	checkStringFixed(got, 2, "8.44", t)
	// Negative
	number = decimal.NewFromInt(19876)
	got, err = bot.CalculateMACD(number, exampleCandles)
	if err != nil {
		t.Error(err)
	}
	checkStringFixed(got, 2, "-5.36", t)
	// Zeroish 🤷🏼‍♂️
	number = decimal.NewFromInt(20031)
	got, err = bot.CalculateMACD(number, exampleCandles)
	if err != nil {
		t.Error(err)
	}
	checkStringFixed(got, 2, "7.00", t)
}

func TestCalculateSignalLine(t *testing.T) {
	// No results with less than 9
	data := []decimal.Decimal{td(1), td(2), td(3)}
	got, err := bot.CalculateSignalLine(data, decimal.Zero)
	if err != bot.ErrCalcSignalNotEnoughInfo {
		t.Error(got, err)
	}
	// Correct at exactly 9
	data = append(data, td(4), td(5), td(6), td(7), td(8), td(9))
	got, _ = bot.CalculateSignalLine(data, decimal.Zero)
	checkStringFixed(got, 2, "5.80", t)
	// Correct past 9
	data = append(data, td(10), td(27), td(3))
	got, _ = bot.CalculateSignalLine(data, decimal.NewFromFloat(10.712))
	checkStringFixed(got, 2, "9.17", t)
}

func TestCalculateHistogram(t *testing.T) {
	got := bot.CalculateHistogram(td(1), td(1))
	checkStringFixed(got, 0, "0", t)
}
