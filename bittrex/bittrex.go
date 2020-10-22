package bittrex

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	aPIVersion = "3"
)

var (
	httpClient = http.Client{
		Timeout: time.Second * 30,
	}
	creds = func() Auth {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
		auth := Auth{
			apiKey:    os.Getenv("BIT_KEY"),
			secretKey: os.Getenv("BIT_SECRET"),
		}
		return auth
	}()
)

// PokeAPI returns any errors the api throws; nil if the API responds with 0 errors
func PokeAPI() error {
	var pingResponse pingResponse
	timestamp := time.Now().UTC().Unix()
	URL := fmt.Sprintf("https://socket.bittrex.com/signalr/ping?_=%d", timestamp)
	response, err := httpClient.Get(URL)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("Status Code: %d", response.StatusCode)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &pingResponse)
	if err != nil {
		return err
	}
	if pingResponse.Response == "pong" {
		return nil
	}

	return errors.New("💩 something is wrong")
}

func get(url string, authenticate bool) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "cryptofu")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Add("Cache-Control", "no-store")
	req.Header.Add("Cache-Control", "must-revalidate")

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

// TODO func post()

func addAuthHeaders(req *http.Request) {
	timestamp := makeTimestamp()
	hash := makeHash("")
	req.Header.Set("Api-Key", creds.apiKey)
	req.Header.Set("Api-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("Api-Content-Hash", hash)
	req.Header.Set("Api-Signature", makeAPISigniture(req, timestamp, hash))
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
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

// GetBitcoin gets the current bitcoin market
func GetBitcoin() (MarketResponse, error) {
	resp, err := get("https://api.bittrex.com/v3/markets/BTC-USD/summary", false)
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

// GetAccount gets your account info
func GetAccount() (AccountResponse, error) {
	resp, err := get("https://api.bittrex.com/v3/account", true)
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
	resp, err := get("https://api.bittrex.com/v3/balances", true)
	if err != nil {
		return BalancesResponce{}, nil
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

// TODO make a buy or sell request
