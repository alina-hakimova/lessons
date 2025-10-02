package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHelloHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(helloHandler)
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Fatalf("Expected status 200, got %d", status)
    }

    expected := "Hello from Go Service!"
    if rr.Body.String() != expected {
        t.Fatalf("Expected body %q, got %q", expected, rr.Body.String())
    }
}

func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/health", nil)
    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(healthHandler)
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Fatalf("Expected status 200, got %d", status)
    }

    expected := `{"status": "healthy"}`
    if rr.Body.String() != expected {
        t.Fatalf("Expected body %q, got %q", expected, rr.Body.String())
    }
}
