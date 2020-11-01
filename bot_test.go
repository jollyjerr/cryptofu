package main

import (
	"cryptofu/bot"
	"testing"
)

func TestCalculateEMASmoothing(t *testing.T) {
	got := bot.CalculateEMASmoothing(6)
	if got.StringFixed(2) != "0.29" {
		t.Errorf("CalculateEMASmoothing was %s, expected 0.29", got.StringFixed(2))
	}
}

// ticket1 := bittrex.TickerResponse{BidRate: "12300"}
// ticket2 := bittrex.TickerResponse{BidRate: "13400"}
// ticket3 := bittrex.TickerResponse{BidRate: "11350"}

// args := []bittrex.TickerResponse{ticket1, ticket2, ticket3}

// sma, err := bot.CalculateSMA(args)
// if err != nil {
// 	fmt.Println(err)
// }
// fmt.Println(sma)

// smooth := bot.CalculateEMASmoothing(3)
// fmt.Println(smooth)

// ticket4, err := decimal.NewFromString("12200")
// if err != nil {
// 	fmt.Println(err)
// }

// ema := bot.CalculateEMA(ticket4, sma, smooth)
// fmt.Println(ema)

// ticket4InTickerForm := bittrex.TickerResponse{BidRate: "12200"}
// tema, err := bot.TickerToTEMA(ticket4InTickerForm, sma, smooth)
// if err != nil {
// 	fmt.Println(err)
// }

// fmt.Println(tema)
