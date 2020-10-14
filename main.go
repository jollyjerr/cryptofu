package main

import (
	"cryptofu/bittrex"
	"fmt"
)

type bot struct {
	mode string
}

func (b bot) Run() {
	fmt.Println("Hello, world")
	api := bittrex.PokeAPI()
	fmt.Println(api)
}

func main() {
	bot := bot{
		mode: "Sandbox",
	}
	bot.Run()
}
