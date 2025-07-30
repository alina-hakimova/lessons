package main

import (
    "context"
    "encoding/json"
    "os"
    "time"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/streadway/amqp"
)

func getEnv(key, defaultValue string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultValue
}

func main() {
    r := gin.Default()
    rabbitHost := getEnv("RABBIT_HOST", "localhost")
    queueName := getEnv("RABBIT_QUEUE", "task_queue")
    port := getEnv("GO_PORT", "5002")

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

    _, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
    if err != nil {
        panic(err)
    }

    r.POST("/tasks", func(c *gin.Context) {
        var data map[string]interface{}
        if err := c.BindJSON(&data); err != nil {
            c.JSON(400, gin.H{"error": "bad request"})
            return
        }
        if _, ok := data["title"]; !ok {
            c.JSON(400, gin.H{"error": "title required"})
            return
        }
        data["created_at"] = time.Now().UTC().Format(time.RFC3339)
        data["status"] = "queued"
        body, _ := json.Marshal(data)

        err = ch.Publish("", queueName, false, false,
            amqp.Publishing{
                DeliveryMode: amqp.Persistent,
                ContentType:  "application/json",
                Body:         body,
            })
        if err != nil {
            c.JSON(500, gin.H{"error": "rabbit publish fail"})
            return
        }
        c.JSON(202, gin.H{"status": "queued", "task": data})
    })

    // (опционально) — если хотите, как у Flask demo:
    r.GET("/tasks", func(c *gin.Context) {
        mongoURI := getEnv("MONGO_URI", "mongodb://mongo:27017/")
        dbname := getEnv("DB_NAME", "mydb")

        cli, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
        if err != nil {
            c.JSON(500, gin.H{"error": "mongo connect fail"})
            return
        }
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        if err := cli.Connect(ctx); err != nil {
            c.JSON(500, gin.H{"error": "mongo connect fail"})
            return
        }
        var results []map[string]interface{}
        cur, err := cli.Database(dbname).Collection("tasks").Find(ctx, map[string]interface{}{})
        if err == nil {
            _ = cur.All(ctx, &results)
        }
        cli.Disconnect(ctx)
        c.JSON(200, results)
    })

    r.Run("0.0.0.0:" + port)
}
