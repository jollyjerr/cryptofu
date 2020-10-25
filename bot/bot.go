package bot

import (
	"cryptofu/bittrex"
	"log"
	"time"

	"go.uber.org/zap"
)

// Bot is the main trading bot
type Bot struct {
	Mode   string
	Symbol string
}

var (
	logger = func() *zap.SugaredLogger {
		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Fatal("💩 Error getting logger set up?")
		}
		defer logger.Sync()
		sugar := logger.Sugar()
		return sugar
	}()
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
	err := bot.SingleRotation(bot.Symbol)
	if err != nil {
		bot.checkErrorAndAct(err)
	} else {
		bot.sleep()
	}
}

// SingleRotation runs the bot trading logic once
func (bot Bot) SingleRotation(symbol string) error {
	logger.Debug("Running a single rotation")
	err := bittrex.PokeAPI()
	if err != nil {
		logger.Error(err)
		return ErrPing
	}
	ticker, err := bittrex.GetTicker(symbol)
	if err != nil {
		logger.Error(err)
		return ErrTicker
	}
	bot.processTickerUpdate(ticker)
	return nil
}

func (bot Bot) checkErrorAndAct(err error) {
	switch err {
	case ErrPing:
		logger.Error("API Ping failed")
		bot.sleep()
	case ErrTicker:
		logger.Error("Failed to get ticker information")
		bot.sleep()
	default:
		SelfDestruct <- true
	}
}

func (bot Bot) sleep() {
	logger.Debug("Sleeping")
	time.Sleep(61 * time.Second)
	bot.Run()
}

func (bot Bot) processTickerUpdate(ticker bittrex.TickerResponse) {
	logger.Info(ticker)
}
