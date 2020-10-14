package bittrex

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// Auth is the type for bittrex creds
type Auth struct {
	apiKey    string
	secretKey string
}

type pingResponse struct {
	Response string `json:"response,required"`
}

// BTCPrice represents the BTC price at the specified timestamp.
type BTCPrice struct {
	USDValue  decimal.Decimal
	Timestamp time.Time
}

type btcPriceResult struct {
	Bpi struct {
		USD struct {
			Code        string      `json:"code,required"`
			Description string      `json:"description,required"`
			Rate        string      `json:"rate,required"`
			RateFloat   json.Number `json:"rate_float,required"`
		} `json:"USD,required"`
		Disclaimer string `json:"disclaimer,required"`
	} `json:"bpi,required"`
	Time struct {
		Updated    string `json:"updated,required"`
		UpdatedISO string `json:"updatedISO,omitempty"`
		UpdatedUK  string `json:"updateduk,omitempty"`
	} `json:"time,required"`
}

func (result btcPriceResult) Compress() BTCPrice {
	value, _ := result.Bpi.USD.RateFloat.Float64()
	ts, _ := time.Parse(time.RFC3339, result.Time.UpdatedISO)
	return BTCPrice{
		USDValue:  decimal.NewFromFloat(value),
		Timestamp: ts,
	}
}
