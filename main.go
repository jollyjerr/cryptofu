package main

import (
	"cryptofu/bittrex"
	"fmt"
	"log"
)

type bot struct {
	mode string
}

func (b bot) Run() {
	fmt.Println("Hello, world")
	api := bittrex.PokeAPI()
	fmt.Println(api)
	res, err := bittrex.Get("https://api.bittrex.com/v3/markets", false)
	if err != nil {
		log.Fatal("uh oh")
	}
	fmt.Println(res)
}

func main() {
	bot := bot{
		mode: "Sandbox",
	}
	bot.Run()
}
