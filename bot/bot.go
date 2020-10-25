package bot

import (
	"cryptofu/bittrex"
	"fmt"
	"log"
	"time"
)

// Bot is the main trading bot
type Bot struct {
	Mode string
}

var (
	// SelfDestruct is a channel the bot can use to kill the go process at any time
	SelfDestruct = make(chan bool)
	// Modes are accepted bot modes
	Modes = map[string]string{
		"Testing":     "testing",
		"Development": "development",
		"Sandbox":     "sandbox",
		"Production":  "production",
	}
)

// Run runs the trading bot
func (bot Bot) Run() {
	err := bot.runRotation()
	if err != nil {
		bot.checkErrorAndAct(err)
	} else {
		bot.sleep()
		bot.Run()
	}
}

func (bot Bot) runRotation() error {
	err := bittrex.PokeAPI()
	if err != nil {
		return ErrPing
	}
	res, err := bittrex.GetTicker(bittrex.Symbols["Bitcoin"])
	if err != nil {
		return err
	}
	fmt.Println("Current bitcoin ticker")
	fmt.Println(res.AskRate)
	fmt.Println(res.BidRate)
	fmt.Println(res.LastTradeRate)
	return nil
}

func (bot Bot) checkErrorAndAct(err error) {
	switch err {
	case ErrPing:
		log.Println("API Ping failed")
		bot.sleep()
		bot.Run()
	default:
		SelfDestruct <- true
	}
}

func (bot Bot) sleep() {
	log.Println("Sleeping")
	time.Sleep(61 * time.Second)
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
