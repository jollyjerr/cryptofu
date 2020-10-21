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
	apierr := bittrex.PokeAPI()
	if apierr != nil {
		log.Fatal("uhhhh ohhhhh")
	}
	res, err := bittrex.GetBitcoin()
	if err != nil {
		log.Fatal("uh oh")
	}
	fmt.Println("The brice of bitcoin is")
	fmt.Println(res.High)

	nres, err := bittrex.GetAccount()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Account ID is....")
	fmt.Println(nres.AccountID)
}

var exit = make(chan bool)

func main() {
	bot := bot{
		mode: "Sandbox",
	}
	go bot.Run()
	<-exit
	fmt.Println("Cryptofu shutting down")
}
