package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

// logRequestTime оборачивает обработчик и логирует время выполнения запроса
func logRequestTime(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        handler(w, r)
        duration := time.Since(start).Milliseconds()
        fmt.Printf("%s %s took %d ms\n", r.Method, r.URL.Path, duration)
    }
}

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

var (
	pgDB           *sql.DB
	mongoClient    *mongo.Client
	taskCollection *mongo.Collection
	redisClient    *redis.Client
)

func main() {
	// Подключение к Postgres
	pgConnStr := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
	)
	var err error
	pgDB, err = sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pgDB.Close()

	// Проверка соединения с Postgres
	err = pgDB.Ping()
	if err != nil {
		log.Fatal("Postgres ping failed:", err)
	}

	// Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(ctx)

	// Проверка соединения с MongoDB
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	taskCollection = mongoClient.Database(os.Getenv("MONGO_DB")).Collection("tasks")

	// Подключение к Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "", // нет пароля
		DB:       0,  // используем DB по умолчанию
	})

	// Проверка соединения с Redis
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Redis ping failed:", err)
	}

	// Роутинг
	http.HandleFunc("/notes", logRequestTime(notesHandler))
        http.HandleFunc("/tasks", logRequestTime(tasksHandler))


	fmt.Println("Server started at :5002")
	log.Fatal(http.ListenAndServe(":5002", nil))
}

/////////////////////////
// NOTES - Postgres
/////////////////////////

func notesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createNoteHandler(w, r)
	case http.MethodGet:
		getNotesHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createNoteHandler(w http.ResponseWriter, r *http.Request) {
	var n Note
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := pgDB.QueryRow(
		"INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id, created_at",
		n.Title, n.Content,
	).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Инвалидация кэша
	redisClient.Del(r.Context(), "notes_list")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(n)
}

func getNotesHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем кэш
	cached, err := redisClient.Get(r.Context(), "notes_list").Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	// Если нет в кэше, идем в БД
	rows, err := pgDB.Query("SELECT id, title, content, created_at FROM notes ORDER BY id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		notes = append(notes, n)
	}

	// Сохраняем в кэш
	jsonData, err := json.Marshal(notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redisClient.SetEx(r.Context(), "notes_list", jsonData, 60*time.Second)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

/////////////////////////
// TASKS - MongoDB
/////////////////////////

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createTaskHandler(w, r)
	case http.MethodGet:
		getTasksHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t.CreatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

res, err := taskCollection.InsertOne(ctx, bson.M{
    "title":       t.Title,
    "description": t.Description,
    "created_at":  t.CreatedAt,
})
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}

objID, ok := res.InsertedID.(primitive.ObjectID)
if ok {
    t.ID = objID.Hex()
} else {
    t.ID = fmt.Sprintf("%v", res.InsertedID)
}

// Инвалидация кэша
redisClient.Del(r.Context(), "tasks_list")

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(t)
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем кэш
	cached, err := redisClient.Get(r.Context(), "tasks_list").Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	// Если нет в кэше, идем в БД
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := taskCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var tasks []Task
	for cursor.Next(ctx) {
		var doc struct {
			ID          string    `bson:"_id"`
			Title       string    `bson:"title"`
			Description string    `bson:"description"`
			CreatedAt   time.Time `bson:"created_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, Task{
			ID:          doc.ID,
			Title:       doc.Title,
			Description: doc.Description,
			CreatedAt:   doc.CreatedAt,
		})
	}

	// Сохраняем в кэш
	jsonData, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
redisClient.SetEx(r.Context(), "tasks_list", jsonData, 60*time.Second)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}
