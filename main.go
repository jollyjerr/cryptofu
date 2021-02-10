package main

import (
	"cryptofu/bittrex"
	"cryptofu/bot"
	"fmt"
)

func main() {
	go bot.NewBot(bot.Modes["Paper"], bittrex.Symbols["Doge"])
	<-bot.SelfDestruct
	fmt.Println("😲 Cryptofu shutting down! 🧨 💥")
}
