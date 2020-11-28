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

// TickerResponse is a ticker response
type TickerResponse struct {
	Symbol        string
	LastTradeRate string
	BidRate       string
	AskRate       string
}

// CandleResponse is a candle response
type CandleResponse struct {
	StartsAt    string
	Open        string
	High        string
	Low         string
	Close       string
	Volume      string
	QuoteVolume string
}

// AccountResponse is an account response
type AccountResponse struct {
	SubAccountID string
	AccountID    string
}

// BalanceResponse is one account ballance
type BalanceResponse struct {
	CurrencySymbol string
	Total          int
	Available      int
	UpdatedAt      string
}

// BalancesResponce is all account ballances
type BalancesResponce []BalanceResponse

// NewOrder is the request body of a new order
type NewOrder struct {
	MarketSymbol  string
	Direction     string
	Type          string
	Quantity      int
	Ceiling       int
	Limit         int
	TimeInForce   string
	ClientOrderID string
	UseAwards     bool
}

// OrderResponse is the response from a new order
type OrderResponse struct {
	ID            string
	MarketSymbol  string
	Direction     string
	Type          string
	Quantity      string
	Limit         string
	Ceiling       string
	TimeInForce   string
	ClientOrderID string
	FillQuantity  string
	Commission    string
	Proceeds      string
	Status        string
	CreatedAt     string
	UpdatedAt     string
	ClosedAt      string
	OrderToCancel struct {
		Type string
		ID   string
	}
}
