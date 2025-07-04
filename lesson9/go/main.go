package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    _ "github.com/lib/pq"
)

type Note struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB

func main() {
    var err error
    db, err = sql.Open("postgres", "host=localhost dbname=notes_db user=notes_user password=mystrongpassword sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    http.HandleFunc("/notes", notesHandler)
    fmt.Println("Server started at :5002")
    log.Fatal(http.ListenAndServe(":5002", nil))
}

func notesHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        createNoteHandler(w, r)
    } else if r.Method == http.MethodGet {
        getNotesHandler(w, r)
    } else {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func createNoteHandler(w http.ResponseWriter, r *http.Request) {
    var n Note
    err := json.NewDecoder(r.Body).Decode(&n)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err = db.QueryRow(
        "INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id, created_at",
        n.Title, n.Content,
    ).Scan(&n.ID, &n.CreatedAt)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(n)
}

func getNotesHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, title, content, created_at FROM notes ORDER BY id")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    notes := []Note{}
    for rows.Next() {
        var n Note
        err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        notes = append(notes, n)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(notes)
}
