package bittrex

// Auth is the type for bittrex creds
type Auth struct {
	apiKey    string
	secretKey string
}

type pingResponse struct {
	Response string `json:"response,required"`
}

// MarketResponse is a market response
type MarketResponse struct {
	Symbol        string
	High          string
	Low           string
	Volume        string
	QuoteVolume   string
	PercentChange string
	UpdatedAt     string
}

// AccountResponse is an account response
type AccountResponse struct {
	SubAccountID string
	AccountID    string
}
