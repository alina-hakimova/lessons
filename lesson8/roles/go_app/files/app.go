package main

import (
	"fmt"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("SECRET_MESSAGE")
	if secret == "" {
		secret = "no secret provided"
	}
	fmt.Fprintf(w, "Secret message from Go: %s\n", secret)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":5000", nil)
}