package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name, _ := os.Hostname()
		fmt.Fprintf(w, "Hello from %s\n", name)
	})
	http.ListenAndServe(":8080", nil)
}
