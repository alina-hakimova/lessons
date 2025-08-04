package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "time"
    "github.com/streadway/amqp"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func getEnv(key, defaultValue string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultValue
}

func main() {
    rabbitHost := getEnv("RABBIT_HOST", "rabbitmq")
    queueName := getEnv("RABBIT_QUEUE", "task_queue")
    mongoURI := getEnv("MONGO_URI", "mongodb://mongo:27017/")
    dbName := getEnv("DB_NAME", "mydb")

    rabbitURL := "amqp://guest:guest@" + rabbitHost + ":5672/"
    conn, err := amqp.Dial(rabbitURL)
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    ch, err := conn.Channel()
    if err != nil {
        panic(err)
    }
    defer ch.Close()
    q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
    if err != nil {
        panic(err)
    }

    msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
    if err != nil {
        panic(err)
    }

    cli, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
    if err != nil {
        panic(err)
    }
    ctx := context.Background()
    if err := cli.Connect(ctx); err != nil {
        panic(err)
    }
    coll := cli.Database(dbName).Collection("tasks")

    for d := range msgs {
        var task map[string]interface{}
        if err := json.Unmarshal(d.Body, &task); err == nil {
            task["processed_at"] = time.Now().UTC().Format(time.RFC3339)
            task["status"] = "completed"
            _, err = coll.InsertOne(ctx, task)
            if err == nil {
                fmt.Printf(" [x] Saved: %v\n", task["title"])
                d.Ack(false)
            } else {
                fmt.Printf(" [x] Mongo insert error: %v\n", err)
            }
        } else {
            fmt.Printf(" [x] JSON error: %v\n", err)
        }
    }
}
