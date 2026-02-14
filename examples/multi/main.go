package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	http.HandleFunc("/api/v1/greeting", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "hello from api-server"})
	})
	http.ListenAndServe(":8080", nil)
}
