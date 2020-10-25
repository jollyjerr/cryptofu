package main

import (
	"cryptofu/bittrex"
	"cryptofu/bot"
	"fmt"
)

func main() {
	cryptofu := bot.Bot{
		Mode:   bot.Modes["Development"],
		Symbol: bittrex.Symbols["Bitcoin"],
	}
	go cryptofu.Run()
	<-bot.SelfDestruct
	fmt.Println("😲 Cryptofu shutting down! 🧨 💥")
	// TODO send an email if the bot exits
}
