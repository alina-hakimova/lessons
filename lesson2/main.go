package main
 import (
 "fmt"
 "net/http"
 )
 func handler(w http.ResponseWriter, r *http.Request) {
 fmt.Fprintf(w, "Hello from Go!")
 }
 func main() {
 http.HandleFunc("/", handler)
 http.ListenAndServe(":8003", nil)
 }
