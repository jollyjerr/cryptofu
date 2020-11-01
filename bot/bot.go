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
	√4. Find EMA for current ticker === ticker.Bid * smoothing + pevTickEMA or SMA * (1 - smoothing)
	√5. Find EMA2 for current ticker === EMA * smoothing + pevTickEMA or SMA * (1 - smoothing)
	√6. Find EMA3 for current ticker === EMA2 * smoothing + pevTickEMA or SMA * (1 - smoothing)
	√7. Find TEMA === (3 * EMA) - (3 * EMA2) + EMA3
	√8. Store TEMA per tick
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
	Mode             string
	Symbol           string
	sleepSeconds     int
	period           int
	sma              decimal.Decimal
	useSma           bool
	tickerHistory    []bittrex.TickerResponse
	temaHistory      []decimal.Decimal
	macdHistory      []decimal.Decimal
	maxHistoryLength int
}

// NewBot makes a new trading bot with very sensible default values
func NewBot(mode string, symbol string) *Bot {
	return &Bot{
		Mode:             mode,
		Symbol:           symbol,
		sleepSeconds:     60,
		period:           10,
		sma:              decimal.NewFromInt(0),
		useSma:           true,
		tickerHistory:    make([]bittrex.TickerResponse, 0),
		temaHistory:      make([]decimal.Decimal, 0),
		macdHistory:      make([]decimal.Decimal, 0),
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
	logger.Debugf("Running a single rotation %d", len(bot.tickerHistory))

	// Check that api is alive
	err := bittrex.PokeAPI()
	if err != nil {
		logger.Error(err)
		return ErrPing
	}

	// Get current ticker of whatever symbol is being tracked
	ticker, err := bittrex.GetTicker(symbol)
	if err != nil {
		return ErrTicker
	}

	// Process that ticker and convert it into useful stats
	err = bot.processTickerUpdate(ticker)
	if err != nil {
		return err // TODO gracefully handle this error
	}

	// Calculate the macd on the current symbol
	err = bot.checkMACD()
	if err != nil {
		return err
	}

	bot.cleanHistory()
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
	case ErrCalcMACDNotEnoughInfo:
		logger.Info("😴 Not enough info to calculate MACD.")
		bot.sleep()
	default:
		logger.Error(err)
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
		logger.Debug("Cleaning oldest ticker record")
		bot.tickerHistory = append(bot.tickerHistory[:0], bot.tickerHistory[1:]...)
	}
	if len(bot.temaHistory) > bot.maxHistoryLength {
		logger.Debug("Cleaning oldest tema record")
		bot.temaHistory = append(bot.temaHistory[:0], bot.temaHistory[1:]...)
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
			tema, err := TickerToTEMA(ticker, bot.sma, bot.smoothingModifier())
			if err != nil {
				return err
			}
			logger.Debug(ticker, tema)
			bot.temaHistory = append(bot.temaHistory, tema)
			// the sma has served it's time
			bot.useSma = false
		} else {
			tema, err := TickerToTEMA(ticker, bot.temaHistory[len(bot.temaHistory)-1], bot.smoothingModifier())
			if err != nil {
				return err
			}
			logger.Debug("Current TEMA:", tema)
			bot.temaHistory = append(bot.temaHistory, tema)
		}
	}
	return nil
}

func (bot *Bot) updateSMA() error {
	if bot.useSma {
		if len(bot.tickerHistory) >= bot.period {
			num, err := CalculateSMA(bot.tickerHistory)
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
	return CalculateEMASmoothing(bot.period)
}

func (bot *Bot) checkMACD() error {
	mostRecentValue, err := decimal.NewFromString(bot.tickerHistory[len(bot.tickerHistory)-1].BidRate)
	if err != nil {
		return err
	}
	macd, err := CalculateMACD(mostRecentValue, bot.tickerHistory)
	if err != nil {
		return err
	}
	logger.Infof("MACD is %s for this tema value: %s", macd.StringFixed(4), mostRecentValue.StringFixed(2))
	bot.macdHistory = append(bot.macdHistory, macd)
	return nil
}
