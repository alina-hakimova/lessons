package main

import (
        "fmt"
        "net/http"
        "os"
)

func versionHandler(w http.ResponseWriter, r *http.Request) {
        version := os.Getenv("VERSION")
        if version == "" {
                version = "dev"
        }
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, "{\"version\": \"%s\"}", version)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, "{\"status\": \"healthy\", \"service\": \"shipments\"}")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, "{\"message\": \"Shipments Service v1\"}")
}

func main() {
        http.HandleFunc("/version", versionHandler)
        http.HandleFunc("/health", healthHandler)
        http.HandleFunc("/", rootHandler)

        port := os.Getenv("PORT")
        if port == "" {
                port = ":8080"
        } else {
                port = ":" + port
        }
        
        fmt.Printf("Server starting on port %s\n", port)
        http.ListenAndServe(port, nil)
}
