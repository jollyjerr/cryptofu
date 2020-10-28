package bot

import (
	"cryptofu/bittrex"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

/*
	√1. Collect data points for a "period" of time
	√2. Sum of the BID prices for the period / number of data points === SMA
	√3. Calculate the smoothing modifier === 2 ÷ (number of observations + 1)
	4. Find EMA for current ticker === ticker.Bid * smoothing + pevTickEMA or SMA * (1 - smoothing)
	5. Find EMA2 for current ticker === EMA * smoothing + pevTickEMA or SMA * (1 - smoothing)
	6. Find EMA3 for current ticker === EMA2 * smoothing + pevTickEMA or SMA * (1 - smoothing)
	7. Find TEMA === (3 * EMA) - (3 * EMA2) + EMA3
	8. Store TEMA per tick
	9. If TEMA is higher than prev TEMA for "x" number of times or at "x" percent increase buy
	10. Activate trailing sell at "x", increase at 1:1 with new TEMA updates. DONT decrese
	11. If TEMA dips below X, sell
*/

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

// Bot is the main trading bot
type Bot struct {
	Mode   string
	Symbol string
	// How big the slices are
	sleepSeconds int
	// How many slices
	period           int
	sma              decimal.Decimal
	useSma           bool
	tickerHistory    []bittrex.TickerResponse
	temaHistory      []decimal.Decimal
	maxHistoryLength int
}

// NewBot makes a new trading bot with very sensible default values
func NewBot(mode string, symbol string) *Bot {
	return &Bot{
		Mode:             mode,
		Symbol:           symbol,
		sleepSeconds:     60,
		period:           2,
		sma:              decimal.NewFromInt(0),
		useSma:           true,
		tickerHistory:    make([]bittrex.TickerResponse, 0),
		temaHistory:      make([]decimal.Decimal, 0),
		maxHistoryLength: 1000, // TODO replace with database or flat files? https://github.com/mongodb/mongo-go-driver
	}
}

// Run runs the trading bot
func (bot *Bot) Run() {
	err := bot.SingleRotation(bot.Symbol)
	if err != nil {
		bot.checkErrorAndAct(err)
	} else {
		bot.sleep()
	}
}

// SingleRotation runs the bot trading logic once
func (bot *Bot) SingleRotation(symbol string) error {
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
	err = bot.processTickerUpdate(ticker)
	if err != nil {
		logger.Error(err)
		return err // TODO gracefully handle this error
	}
	return nil
}

func (bot *Bot) checkErrorAndAct(err error) {
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

func (bot *Bot) sleep() {
	if bot.Mode == Modes["Development"] || bot.Mode == Modes["Production"] {
		logger.Debug("Sleeping")
		time.Sleep(time.Duration(bot.sleepSeconds) * time.Second)
	} else {
		logger.Debug("Starting next cycle")
	}
	bot.Run()
}

func (bot *Bot) cleanHistory() {
	if len(bot.tickerHistory) > bot.maxHistoryLength {
		logger.Debug("Cleaning oldest history record")
		bot.tickerHistory = append(bot.tickerHistory[:0], bot.tickerHistory[1:]...)
	}
}

func (bot *Bot) processTickerUpdate(ticker bittrex.TickerResponse) error {
	logger.Debug(ticker)
	bot.tickerHistory = append(bot.tickerHistory, ticker)
	err := bot.updateSMA()
	if err != nil {
		return err
	}
	if bot.sma.IsZero() {
		logger.Infof("😴 Not enough info to make a calculation. %d out of %d needed cycles", len(bot.tickerHistory), bot.period)
	} else {
		if bot.useSma {
			tema, err := tickerToTEMA(ticker, bot.sma)
			if err != nil {
				return err
			}
			logger.Debug(ticker, tema)
			bot.temaHistory = append(bot.temaHistory, tema)
			// the sma has served it's time
			bot.useSma = false
		} else {
			tema, err := tickerToTEMA(ticker, bot.temaHistory[len(bot.temaHistory)-1])
			if err != nil {
				return err
			}
			logger.Debug(ticker, tema)
			bot.temaHistory = append(bot.temaHistory, tema)
		}
	}
	bot.cleanHistory()
	return nil
}

func (bot *Bot) updateSMA() error {
	if bot.useSma {
		if len(bot.tickerHistory) >= bot.period {
			num, err := calculateSMA(bot.tickerHistory)
			if err != nil {
				return err
			}
			logger.Infof("🎉🎉🎉 Updating SMA to %s", num)
			bot.sma = num
		}
	}
	return nil
}

func (bot *Bot) smoothingModifier() decimal.Decimal {
	return decimal.NewFromInt(2).Div(decimal.NewFromInt(int64(bot.period + 1)))
}
