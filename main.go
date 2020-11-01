package main

import (
	"cryptofu/bittrex"
	"cryptofu/bot"
	"fmt"
)

func main() {
	// cryptofu := bot.NewBot(bot.Modes["Development"], bittrex.Symbols["Bitcoin"])
	// go cryptofu.Run()
	// <-bot.SelfDestruct
	// fmt.Println("😲 Cryptofu shutting down! 🧨 💥")

	ticket1 := bittrex.TickerResponse{BidRate: "12300"}
	ticket2 := bittrex.TickerResponse{BidRate: "13400"}
	ticket3 := bittrex.TickerResponse{BidRate: "11350"}

	args := []bittrex.TickerResponse{ticket1, ticket2, ticket3}

	sma, err := bot.CalculateSMA(args)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(sma)
}
