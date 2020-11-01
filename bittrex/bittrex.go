package bittrex

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	baseURL    = "https://api.bittrex.com"
	aPIVersion = "v3"
)

var (
	httpClient = http.Client{
		Timeout: time.Second * 30,
	}
	creds = func() Auth {
		if os.Getenv("TEST") == "true" {
			return Auth{}
		}
		err := godotenv.Load()
		if err != nil {
			log.Fatal("💩 Error loading .env file")
		}
		auth := Auth{
			apiKey:    os.Getenv("BIT_KEY"),
			secretKey: os.Getenv("BIT_SECRET"),
		}
		return auth
	}()
	// Symbols is a stored association of market symbols to paramatize args
	Symbols = map[string]string{
		"Bitcoin": "BTC-USD",
	}
	// CandleIntervals is a stored association of candle intervals to paramatize args
	CandleIntervals = map[string]string{
		"1min":  "MINUTE_1",
		"5min":  "MINUTE_5",
		"1hour": "HOUR_1",
		"1day":  "DAY_1",
	}
)

// PokeAPI returns any errors the api throws; nil if the API responds with 0 errors
func PokeAPI() error {
	url := fmt.Sprintf("%s/%s/ping", baseURL, aPIVersion)
	response, err := httpClient.Get(url)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("Status Code: %d", response.StatusCode)
	}

	return nil
}

func get(url string, authenticate bool) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	addStandardHeaders(req)
	if authenticate {
		addAuthHeaders(req)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status Code: %d", resp.StatusCode)
	}

	return resp, nil
}

func post(url string, authenticate bool, body interface{}) (*http.Response, error) {
	marshaledBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshaledBody))
	if err != nil {
		return nil, err
	}

	addStandardHeaders(req)
	if authenticate {
		addAuthHeaders(req)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status Code: %d", resp.StatusCode)
	}

	return resp, nil
}

func addStandardHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "cryptofu")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Add("Cache-Control", "no-store")
	req.Header.Add("Cache-Control", "must-revalidate")
}

func addAuthHeaders(req *http.Request) {
	timestamp := makeTimestamp()
	hash := makeHash("")
	req.Header.Set("Api-Key", creds.apiKey)
	req.Header.Set("Api-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("Api-Content-Hash", hash)
	req.Header.Set("Api-Signature", makeAPISigniture(req, timestamp, hash))
}

func makeTimestamp() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Millisecond)
}

func makeHash(value string) string {
	hasher := sha512.New()
	hasher.Write([]byte(value))
	return hex.EncodeToString(hasher.Sum(nil))
}

func makeAPISigniture(req *http.Request, timestamp int64, contentHash string) string {
	// Empty string for subaccount ID https://bittrex.github.io/api/v3#api-signature
	preSign := fmt.Sprintf("%d%s%s%s%s", timestamp, req.URL, req.Method, contentHash, "")

	hasher := hmac.New(sha512.New, []byte(creds.secretKey))
	hasher.Write([]byte(preSign))

	return hex.EncodeToString(hasher.Sum(nil))
}

// GetMarket gets the daily market values for a symbol
func GetMarket(symbol string) (MarketResponse, error) {
	url := fmt.Sprintf("%s/%s/markets/%s/summary", baseURL, aPIVersion, symbol)
	resp, err := get(url, false)
	if err != nil {
		return MarketResponse{}, err
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return MarketResponse{}, err
	}

	var ret MarketResponse
	err = json.Unmarshal(content, &ret)
	if err != nil {
		return MarketResponse{}, err
	}

	return ret, nil
}

// GetTicker gets the current market ticker for a symbol
func GetTicker(symbol string) (TickerResponse, error) {
	url := fmt.Sprintf("%s/%s/markets/%s/ticker", baseURL, aPIVersion, symbol)
	resp, err := get(url, false)
	if err != nil {
		return TickerResponse{}, err
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TickerResponse{}, err
	}

	var ret TickerResponse
	err = json.Unmarshal(content, &ret)
	if err != nil {
		return TickerResponse{}, err
	}

	return ret, nil
}

// GetCandle gets recent candles for a specific market and interval
func GetCandle(symbol string, interval string) ([]CandleResponse, error) {
	defaultRes := make([]CandleResponse, 0)
	url := fmt.Sprintf("%s/%s/markets/%s/candles/%s/recent", baseURL, aPIVersion, symbol, interval)
	resp, err := get(url, false)
	if err != nil {
		return defaultRes, err
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return defaultRes, err
	}

	var ret []CandleResponse
	err = json.Unmarshal(content, &ret)
	if err != nil {
		return defaultRes, err
	}

	return ret, nil
}

// GetAccount gets your account info
func GetAccount() (AccountResponse, error) {
	url := fmt.Sprintf("%s/%s/account", baseURL, aPIVersion)
	resp, err := get(url, true)
	if err != nil {
		return AccountResponse{}, err
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return AccountResponse{}, err
	}

	var ret AccountResponse
	err = json.Unmarshal(content, &ret)
	if err != nil {
		return AccountResponse{}, err
	}

	return ret, nil
}

// GetBalances gets the balances of all currencies in your account
func GetBalances() (BalancesResponce, error) {
	url := fmt.Sprintf("%s/%s/balances", baseURL, aPIVersion)
	resp, err := get(url, true)
	if err != nil {
		return BalancesResponce{}, err
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BalancesResponce{}, err
	}

	var ret BalancesResponce
	err = json.Unmarshal(content, &ret)
	if err != nil {
		return BalancesResponce{}, err
	}

	return ret, nil
}

// Order requests a new order
func Order(orderDetails NewOrder) (OrderResponse, error) {
	url := fmt.Sprintf("%s/%s/orders", baseURL, aPIVersion)
	resp, err := post(url, true, orderDetails)
	if err != nil {
		return OrderResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return OrderResponse{}, fmt.Errorf("Status Code: %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return OrderResponse{}, err
	}

	var ret OrderResponse
	err = json.Unmarshal(content, &ret)
	if err != nil {
		return OrderResponse{}, err
	}

	return ret, nil
}
