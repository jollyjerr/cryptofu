package bot

import (
	"cryptofu/bittrex"
	"fmt"
	"log"
	"time"
)

// SelfDestruct is a channel the bot can use to kill the go process at any time
var SelfDestruct = make(chan bool)

// Bot is the main trading bot
type Bot struct {
	Mode string
}

// Run runs the trading bot
func (b Bot) Run() {
	doTheThing()
	checkTicker()
	time.Sleep(8 * time.Second)
	checkTicker()
	time.Sleep(8 * time.Second)
	checkTicker()

	SelfDestruct <- true
}

func checkTicker() {
	res, err := bittrex.GetTicker(bittrex.Symbols["Bitcoin"])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Current bitcoin ticker")
	fmt.Println(res.AskRate)
	fmt.Println(res.BidRate)
	fmt.Println(res.LastTradeRate)
}

func doTheThing() {
	fmt.Println("Hello, world")
	apierr := bittrex.PokeAPI()
	if apierr != nil {
		log.Fatal(apierr)
	}
	res, err := bittrex.GetMarket(bittrex.Symbols["Bitcoin"])
	if err != nil {
		log.Fatal("uh oh")
	}
	fmt.Println("The market high of bitcoin is")
	fmt.Println(res.High)

	nres, err := bittrex.GetAccount()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Account ID is....")
	fmt.Println(nres.AccountID)

	nnres, err := bittrex.GetBalances()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Balances are....")
	fmt.Println(nnres)
}
