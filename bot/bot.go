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
		"Testing":    "testing",
		"Production": "production",
	}
	intervalToSleepSeconds = map[string]int{
		bittrex.CandleIntervals["1min"]: 70,
		bittrex.CandleIntervals["5min"]: 300,
	}
	intervalToPeriod = map[string]int{
		bittrex.CandleIntervals["1min"]: 1,
		bittrex.CandleIntervals["5min"]: 1,
	}
)

// Bot is the main trading bot
type Bot struct {
	Mode             string
	Symbol           string
	Interval         string
	Period           int
	trailLag         decimal.Decimal
	candleHistory    []bittrex.CandleResponse
	temaHistory      []decimal.Decimal
	macdHistory      []decimal.Decimal
	signalHistory    []decimal.Decimal
	orderHistory     []bittrex.OrderResponse
	maxHistoryLength int
	currentOrder     bittrex.OrderResponse
	currentTrail     decimal.Decimal
}

// NewBot makes a new trading bot with very sensible default values
func NewBot(mode string, symbol string) *Bot {
	babyBot := Bot{
		Mode:             mode,
		Symbol:           symbol,
		Interval:         bittrex.CandleIntervals["1min"],
		Period:           intervalToPeriod[bittrex.CandleIntervals["1min"]],
		trailLag:         decimal.NewFromInt(1),
		candleHistory:    make([]bittrex.CandleResponse, 0),
		temaHistory:      make([]decimal.Decimal, 0),
		macdHistory:      make([]decimal.Decimal, 0),
		signalHistory:    []decimal.Decimal{decimal.Zero},
		orderHistory:     make([]bittrex.OrderResponse, 0),
		maxHistoryLength: 10000, // TODO replace with database or flat files? Do we even need to? https://github.com/mongodb/mongo-go-driver
		currentOrder:     bittrex.OrderResponse{},
		currentTrail:     decimal.Zero,
	}
	babyBot.Setup()
	return &babyBot
}

// Setup populates a new bot with data and starts the calculations rolling. Errors during this stage are fatal.
func (bot *Bot) Setup() {
	// Point at historical data in testing mode
	if bot.Mode == Modes["Testing"] {
		logger.Info("Running in testing mode: starting fake server")
		bittrex.SetBaseURL("http://localhost:8000")
		go bittrex.StartMockServer()
	}
	bot.SayHi()
	logger.Info("Getting things ready...")
	// Get starting data
	recentCandles, err := bittrex.GetCandles(bot.Symbol, bot.Interval)
	if err != nil {
		logger.Fatal(err)
	}
	// Calculate the sma and first tema based on bot's period
	bot.candleHistory = append(bot.candleHistory, recentCandles[:bot.Period]...)
	sma, err := CandlesToSMA(recentCandles[:bot.Period])
	if err != nil {
		logger.Fatal(err)
	}
	firstTema, err := CandleToTEMA(recentCandles[bot.Period*2], sma, bot.smoothingModifier())
	if err != nil {
		logger.Fatal(err)
	}
	bot.temaHistory = append(bot.temaHistory, firstTema)
	// Calculate the tema for the remaining candles
	remainingCandles := recentCandles[bot.Period*3 : len(recentCandles)-1]
	for i := 0; i < len(remainingCandles); i++ {
		err = bot.processCandleUpdate(remainingCandles[i])
		if err != nil {
			logger.Fatal(err)
		}
	}
	// Go back and populate macd and signal values
	for i := 26; i < len(bot.candleHistory); i++ {
		history := bot.candleHistory[:i]
		val, err := decimal.NewFromString(history[len(history)-1].Close)
		if err != nil {
			logger.Fatal(err)
		}
		macd, err := CalculateMACD(val, history)
		if err == ErrCalcMACDNotEnoughInfo {
			continue
		}
		bot.macdHistory = append(bot.macdHistory, macd)
		err = bot.updateSignal()
		if err != nil && err != ErrCalcSignalNotEnoughInfo {
			logger.Fatal(err)
		}
	}
	// Log startup info
	macd := bot.macdHistory[len(bot.macdHistory)-1]
	signal := bot.signalHistory[len(bot.signalHistory)-1]
	histogram := CalculateHistogram(macd, signal)
	logger.Infof("Starting SMA value was %s", sma.StringFixed(2))
	logger.Infof("Current MACD is: %s, MACD Signal is: %s, and MACD Histogram is: %s", macd.StringFixed(2), signal.StringFixed(2), histogram.StringFixed(2))
	logger.Infof("Bot is ready to go with %d candles processed", len(bot.candleHistory))
	logger.Infof("The last close was %s", bot.candleHistory[len(bot.candleHistory)-1].Close)
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
	logger.Info("Starting rotation")

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

	// Calculate the macd signal line
	err = bot.updateSignal()
	if err != nil {
		return err
	}

	// Decide what to do based on current data
	err = bot.decideRoundAction()
	if err != nil {
		return err
	}

	bot.logRoundStats()
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
		if bot.Mode == Modes["Testing"] {
			logger.Info("Current Order:", bot.currentOrder)
			logger.Info("Current Trail:", bot.currentTrail)
			logger.Info("Order History:", len(bot.orderHistory))
			printStats()
			logger.Info(bot.orderHistory)
			SelfDestruct <- true
		}
		bot.sleep()
	case ErrTicker:
		logger.Error("Failed to get ticker information.")
		bot.sleep()
	case ErrCalcMACDNotEnoughInfo:
		logger.Info("😴 Not enough info to calculate MACD.")
		bot.sleep()
	case ErrCalcSignalNotEnoughInfo:
		logger.Info("😴 Not enough info to calculate Signal.")
		bot.sleep()
	default:
		logger.Error(err)
		SelfDestruct <- true
	}
}

func (bot *Bot) sleep() {
	if bot.Mode == Modes["Production"] {
		logger.Info("Sleeping")
		time.Sleep(time.Duration(intervalToSleepSeconds[bot.Interval]) * time.Second)
	} else {
		logger.Debug("Starting next cycle")
	}
	bot.Run()
}

func (bot *Bot) processCandleUpdate(candle bittrex.CandleResponse) error {
	bot.candleHistory = append(bot.candleHistory, candle)
	tema, err := CandleToTEMA(candle, bot.temaHistory[len(bot.temaHistory)-1], bot.smoothingModifier())
	if err != nil {
		return err
	}
	bot.temaHistory = append(bot.temaHistory, tema)
	return nil
}

func (bot *Bot) processCandlesUpdate(candles []bittrex.CandleResponse) error {
	for i := 1; i < bot.Period+1; i++ {
		err := bot.processCandleUpdate(candles[len(candles)-i])
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
	bot.macdHistory = append(bot.macdHistory, macd)
	return nil
}

func (bot *Bot) updateSignal() error {
	signal, err := CalculateSignalLine(bot.macdHistory, bot.signalHistory[len(bot.signalHistory)-1])
	if err != nil {
		return err
	}
	bot.signalHistory = append(bot.signalHistory, signal)
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
	if len(bot.signalHistory) > bot.maxHistoryLength {
		logger.Debug("Cleaning oldest signal records")
		bot.signalHistory = bot.signalHistory[len(bot.signalHistory)-bot.maxHistoryLength:]
	}
}

func (bot *Bot) decideRoundAction() error {
	var (
		tema           = bot.temaHistory[len(bot.temaHistory)-1]
		macd           = bot.macdHistory[len(bot.macdHistory)-1]
		signal         = bot.signalHistory[len(bot.signalHistory)-1]
		histogram      = CalculateHistogram(macd, signal)
		currentOrderID = bot.currentOrder.ID
	)

	// Update the trail
	if currentOrderID != "" {
		// Update the trail if price has gone up
		if tema.GreaterThan(bot.currentTrail) {
			bot.currentTrail = tema.Sub(bot.trailLag)
			logger.Infof("New trail is at %s", bot.currentTrail.StringFixed(2))
		}

		err := bot.decideShouldSell(tema, histogram, currentOrderID)
		if err != nil {
			return err
		}
	} else {
		err := bot.decideShouldBuy(tema, histogram)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bot *Bot) decideShouldSell(tema decimal.Decimal, histogram decimal.Decimal, currentOrderID string) error {
	close, _ := decimal.NewFromString(bot.candleHistory[len(bot.candleHistory)-1].Close)
	if close.LessThan(bot.currentTrail) && histogram.LessThan(decimal.Zero) { // This is the issue - need to fail faster!!!! But taper this control with the histogram so that it does not fail too fast
		logger.Info("Making a sell")

		copy := bot.currentOrder
		// copy.Direction = "sell"
		copy.ID = bot.candleHistory[len(bot.candleHistory)-1].Close + "candle"
		copy.MarketSymbol = tema.StringFixed(2) + "tema"
		copy.Direction = bot.currentTrail.StringFixed(2) + "trail"
		copy.CreatedAt = bot.candleHistory[len(bot.candleHistory)-1].StartsAt
		copy.OrderToCancel.ID = currentOrderID
		saveSell(bot.candleHistory[len(bot.candleHistory)-1])

		bot.orderHistory = append(bot.orderHistory, copy)
		bot.currentTrail = decimal.Zero
		bot.currentOrder = bittrex.OrderResponse{}
	}
	return nil
}

func (bot *Bot) decideShouldBuy(tema decimal.Decimal, histogram decimal.Decimal) error {
	if histogram.GreaterThan(decimal.NewFromInt(6)) {
		logger.Info("Making a purchase")
		bot.currentTrail = tema.Sub(bot.trailLag)
		bot.currentOrder = bittrex.OrderResponse{ID: bot.candleHistory[len(bot.candleHistory)-1].Close, Direction: "buy", CreatedAt: bot.candleHistory[len(bot.candleHistory)-1].StartsAt}
		bot.orderHistory = append(bot.orderHistory, bot.currentOrder)
		saveBuy(bot.candleHistory[len(bot.candleHistory)-1])
	}
	return nil
}

func (bot *Bot) logRoundStats() {
	candle := bot.candleHistory[len(bot.candleHistory)-1]
	tema := bot.temaHistory[len(bot.temaHistory)-1]
	macd := bot.macdHistory[len(bot.macdHistory)-1]
	signal := bot.signalHistory[len(bot.signalHistory)-1]
	histogram := CalculateHistogram(macd, signal)
	logger.Infof("The latest candle is from %s and closed at %s", candle.StartsAt, candle.Close)
	logger.Infof("The TEMA came out to %s", tema.StringFixed(3))
	logger.Infof("The MACD is %s, with a signal of %s", macd.StringFixed(2), signal.StringFixed(2))
	logger.Infof("The histogram value is %s", histogram.StringFixed(2))
}

// SayHi is a smoke test
func (bot *Bot) SayHi() {
	err := bittrex.PokeAPI()
	if err != nil {
		logger.Fatal("💩", err)
	}
	account, err := bittrex.GetAccount()
	if err != nil {
		logger.Fatal(err)
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
	logger.Infof("👋 Hello account %s!", account.AccountID)
}
