package main

import (
	"cryptofu/bittrex"
	"cryptofu/bot"
	"fmt"
)

func main() {
	cryptofu := bot.NewBot(bot.Modes["Development"], bittrex.Symbols["Bitcoin"])
	go cryptofu.Run()
	<-bot.SelfDestruct
	fmt.Println("😲 Cryptofu shutting down! 🧨 💥")
}
