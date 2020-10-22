package main

import (
	"cryptofu/bot"
	"fmt"
)

func main() {
	cryptofu := bot.Bot{
		Mode: "Sandbox",
	}
	go cryptofu.Run()
	<-bot.SelfDestruct
	fmt.Println("😲 Cryptofu shutting down! 🧨 💥")
	// TODO send an email if the bot exits
}
