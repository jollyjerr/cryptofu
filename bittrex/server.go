package bittrex

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func getPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := pingResponse{Response: "true"}
	json.NewEncoder(w).Encode(response)
}

// Get all books
func getCandles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(books)
}

// StartMockServer starts a new fake bitterex api
func StartMockServer() {
	// Init router
	r := mux.NewRouter()

	// Route handles & endpoints
	r.HandleFunc("/v3/ping", getPing).Methods("GET")
	r.HandleFunc("/books/{id}", getCandles).Methods("GET")

	// Start server
	log.Fatal(http.ListenAndServe(":8000", r))
}
