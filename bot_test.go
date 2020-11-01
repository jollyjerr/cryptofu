package main

import (
	"cryptofu/bittrex"
	"cryptofu/bot"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

var (
	exampleTickers = func() []bittrex.TickerResponse {
		data := make([]bittrex.TickerResponse, 0)
		for i := 0; i < 30; i++ {
			data = append(data, createDemoTickerResponse(i))
		}
		return data
	}()
)

func createDemoTickerResponse(i int) bittrex.TickerResponse {
	return bittrex.TickerResponse{
		BidRate: fmt.Sprintf("%d", 20000+i),
	}
}

func checkStringFixed(forThis decimal.Decimal, fixedAmt int32, expected string, t *testing.T) {
	if forThis.StringFixed(fixedAmt) != expected {
		t.Errorf("Was %s, expected %s", forThis.StringFixed(fixedAmt), expected)
	}
}

/*
	analysis.go
*/

func TestCalculateSMA(t *testing.T) {
	got, err := bot.CalculateSMA(exampleTickers)
	if err != nil {
		t.Error(err)
	}
	checkStringFixed(got, 2, "20014.50", t)
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

func TestTickerToEMA(t *testing.T) {
	// TODO
}

func TestTickerToTEMA(t *testing.T) {
	// TODO
}

func TestCalculateMACD(t *testing.T) {
	// Positive
	number := decimal.NewFromInt(20100)
	got, err := bot.CalculateMACD(number, exampleTickers)
	if err != nil {
		t.Error(err)
	}
	checkStringFixed(got, 2, "1.06", t)
	// Negative
	number = decimal.NewFromInt(19000)
	got, err = bot.CalculateMACD(number, exampleTickers)
	if err != nil {
		t.Error(err)
	}
	checkStringFixed(got, 2, "-86.69", t)
	// Zeroish 🤷🏼‍♂️
	number = decimal.NewFromInt(20087)
	got, err = bot.CalculateMACD(number, exampleTickers)
	if err != nil {
		t.Error(err)
	}
	checkStringFixed(got, 2, "0.02", t)
}
