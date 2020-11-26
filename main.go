package main

import (
	"cryptofu/bittrex"
	"cryptofu/bot"
	"fmt"
)

func main() {
	go bot.NewBot(bot.Modes["Testing"], bittrex.Symbols["Bitcoin"])
	<-bot.SelfDestruct
	fmt.Println("😲 Cryptofu shutting down! 🧨 💥")
}
