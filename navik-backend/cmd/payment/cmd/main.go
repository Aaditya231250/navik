package main

import (
    "encoding/json"
    "net/http"
    "log"
)

type WalletResponse struct {
    Amount int `json:"amount"`
}

func walletHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    resp := WalletResponse{Amount: 0}
    json.NewEncoder(w).Encode(resp)
}

func main() {
    http.HandleFunc("/wallet", walletHandler)
    log.Println("Starting server on :8087")
    log.Fatal(http.ListenAndServe(":8087", nil))
}
