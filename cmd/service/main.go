package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type OrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Service string `json:"service"`
}

func main() {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "unknown-service"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	http.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		// Extract order ID
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 || parts[2] == "" {
			http.Error(w, "missing order id", http.StatusBadRequest)
			return
		}
		orderID := parts[2]

		resp := OrderResponse{
			OrderID: orderID,
			Status:  "shipped",
			Service: serviceName,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Printf("Starting %s on port %s...", serviceName, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
