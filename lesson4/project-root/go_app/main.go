package main

import (
	"fmt"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from Go!")
}

func main() {
	port := os.Getenv("GO_PORT")
	if port == "" {
		port = "8003"
	}
	http.HandleFunc("/", handler)
	fmt.Printf("Go app listening on port %s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}
