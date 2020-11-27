package bittrex

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	candleRequestCount = 0
)

func getPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := pingResponse{Response: "true"}
	json.NewEncoder(w).Encode(response)
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := AccountResponse{SubAccountID: "Mochi and Bao", AccountID: "Test User"}
	json.NewEncoder(w).Encode(response)
}

// Get all books
func getCandles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	candleRequestCount++

	response, err := getCandleResponse(candleRequestCount)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	json.NewEncoder(w).Encode(response)
}

// StartMockServer starts a new fake bitterex api
func StartMockServer() {
	r := mux.NewRouter()

	r.HandleFunc(fmt.Sprintf("/%s/ping", APIVersion), getPing).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/account", APIVersion), getAccount).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/markets/BTC-USD/candles/MINUTE_1/recent", APIVersion), getCandles).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", r))
}
