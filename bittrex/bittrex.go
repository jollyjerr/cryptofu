package bittrex

import (
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

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status Code: %d", resp.StatusCode)
	}

	return resp, nil
}

// GetBitcoin gets the current bitcoin market
func GetBitcoin() (MarketResponse, error) {
	resp, err := get("https://api.bittrex.com/v3/markets/BTC-USD/summary", false)

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
