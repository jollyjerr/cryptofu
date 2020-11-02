package bot

import (
	"cryptofu/bittrex"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

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
	periodToSleepSeconds = map[string]int{
		bittrex.CandleIntervals["1min"]: 60,
		// TODO the rest of this
	}
)

// Bot is the main trading bot
type Bot struct {
	Mode             string
	Symbol           string
	Interval         string
	Period           int
	candleHistory    []bittrex.CandleResponse
	temaHistory      []decimal.Decimal
	macdHistory      []decimal.Decimal
	maxHistoryLength int
}

// NewBot makes a new trading bot with very sensible default values
func NewBot(mode string, symbol string) *Bot {
	babyBot := Bot{
		Mode:             mode,
		Symbol:           symbol,
		Interval:         bittrex.CandleIntervals["1min"],
		Period:           1,
		candleHistory:    make([]bittrex.CandleResponse, 0),
		temaHistory:      make([]decimal.Decimal, 0),
		macdHistory:      make([]decimal.Decimal, 0),
		maxHistoryLength: 1000, // TODO replace with database or flat files? https://github.com/mongodb/mongo-go-driver
	}
	babyBot.Setup()
	return &babyBot
}

// Setup populates a new bot with data and starts the calculations rolling. Errors during this stage are fatal.
func (bot *Bot) Setup() {
	bot.SayHi()
	// Get starting data
	recentCandles, err := bittrex.GetCandles(bot.Symbol, bot.Interval)
	if err != nil {
		logger.Fatal(err)
	}
	// Calculate the sma and first tema based on bot's period
	bot.candleHistory = append(bot.candleHistory, recentCandles[:bot.Period]...)
	sma, err := CalculateSMA(recentCandles[:bot.Period])
	if err != nil {
		logger.Fatal(err)
	}
	firstTema, err := CandleToTEMA(recentCandles[bot.Period+1], sma, bot.smoothingModifier())
	if err != nil {
		logger.Fatal(err)
	}
	bot.temaHistory = append(bot.temaHistory, firstTema)
	// Calculate the tema for the remaining candles
	remainingCandles := recentCandles[bot.Period+2 : len(recentCandles)-1]
	for i := 0; i < len(remainingCandles); i++ {
		err = bot.processCandleUpdate(remainingCandles[i])
		if err != nil {
			logger.Fatal(err)
		}
	}
	// If possible, calculate a starting macd value
	bot.updateMACD()
	// Log startup info
	logger.Infof("Starting SMA value was %s", sma.StringFixed(2))
	logger.Infof("Bot is ready to go with %d candles processed", len(bot.candleHistory))
	bot.sleep()
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
	logger.Debugf("Starting rotation %d", len(bot.candleHistory)+1)

	// Check that api is alive
	err := bittrex.PokeAPI()
	if err != nil {
		logger.Error(err)
		return ErrPing
	}

	// Get current tcandles of whatever symbol is being tracked
	candles, err := bittrex.GetCandles(symbol, bot.Interval)
	if err != nil {
		return ErrCandles
	}

	// Process that ticker and convert it into useful stats
	err = bot.processCandlesUpdate(candles)
	if err != nil {
		return err // TODO gracefully handle this error
	}

	// Calculate the macd on the current symbol
	err = bot.updateMACD()
	if err != nil {
		return err
	}

	bot.cleanHistory()
	return nil
}

func (bot *Bot) checkErrorAndAct(err error) {
	switch err {
	case ErrPing:
		logger.Error("API Ping failed.")
		bot.sleep()
	case ErrCandles:
		logger.Error("Failed to get Candle information.")
		bot.sleep()
	case ErrTicker:
		logger.Error("Failed to get ticker information.")
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
		time.Sleep(time.Duration(periodToSleepSeconds[bot.Interval]) * time.Second)
	} else {
		logger.Debug("Starting next cycle")
	}
	bot.Run()
}

func (bot *Bot) processCandleUpdate(candle bittrex.CandleResponse) error {
	// logger.Debug(candle)
	bot.candleHistory = append(bot.candleHistory, candle)
	tema, err := CandleToTEMA(candle, bot.temaHistory[len(bot.temaHistory)-1], bot.smoothingModifier())
	if err != nil {
		return err
	}
	bot.temaHistory = append(bot.temaHistory, tema)
	return nil
}

func (bot *Bot) processCandlesUpdate(candles []bittrex.CandleResponse) error {
	for i := 0; i < bot.Period; i++ {
		err := bot.processCandleUpdate(candles[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (bot *Bot) smoothingModifier() decimal.Decimal {
	return CalculateEMASmoothing(bot.Period)
}

func (bot *Bot) updateMACD() error {
	mostRecentValue, err := decimal.NewFromString(bot.candleHistory[len(bot.candleHistory)-1].Close)
	if err != nil {
		return err
	}
	macd, err := CalculateMACD(mostRecentValue, bot.candleHistory)
	if err != nil {
		return err
	}
	logger.Infof("MACD is %s for this tema value: %s", macd.StringFixed(4), mostRecentValue.StringFixed(2))
	bot.macdHistory = append(bot.macdHistory, macd)
	return nil
}

func (bot *Bot) cleanHistory() {
	if len(bot.candleHistory) > bot.maxHistoryLength {
		logger.Debug("Cleaning oldest candle records")
		bot.candleHistory = bot.candleHistory[len(bot.candleHistory)-bot.maxHistoryLength:]
	}
	if len(bot.temaHistory) > bot.maxHistoryLength {
		logger.Debug("Cleaning oldest tema records")
		bot.temaHistory = bot.temaHistory[len(bot.temaHistory)-bot.maxHistoryLength:]
	}
	if len(bot.macdHistory) > bot.maxHistoryLength {
		logger.Debug("Cleaning oldest macd records")
		bot.macdHistory = bot.macdHistory[len(bot.macdHistory)-bot.maxHistoryLength:]
	}
}

// SayHi is a smoke test
func (bot *Bot) SayHi() {
	err := bittrex.PokeAPI()
	if err != nil {
		logger.Fatal(err)
	}
	account, err := bittrex.GetAccount()
	if err != nil {
		logger.Fatal()
	}
	message := `
   _____                  _         __       
  / ____|                | |       / _|      
 | |     _ __ _   _ _ __ | |_ ___ | |_ _   _ 
 | |    | '__| | | | '_ \| __/ _ \|  _| | | |
 | |____| |  | |_| | |_) | || (_) | | | |_| |
  \_____|_|   \__, | .__/ \__\___/|_|  \__,_|
               __/ | |                       
              |___/|_|                  
	`
	fmt.Println(message)
	logger.Infof("Hello account %s!", account.AccountID)
}
