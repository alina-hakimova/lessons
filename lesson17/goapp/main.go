package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
  "github.com/gin-contrib/prometheus"
)

func main() {
    r := gin.Default()
    port := getEnv("GO_PORT", "5002")
    p := ginprometheus.NewPrometheus("gin")
    p.Use(r)
    
    // Инициализация RabbitMQ (вынесено в отдельную функцию)
    ch, rabbitConn := initRabbitMQ()
    defer rabbitConn.Close()
    defer ch.Close()

    // Инициализация MongoDB (вынесено в отдельную функцию)
    mongoClient := initMongoDB()
    defer mongoClient.Disconnect(context.Background())

    // Метрики и обработчики остаются без изменений
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    r.POST("/tasks", createTaskHandler(ch))
    r.GET("/tasks", getTasksHandler(mongoClient))
    
    r.Run("0.0.0.0:" + port)
}

// Вынесенные функции:

func initRabbitMQ() (*amqp.Channel, *amqp.Connection) {
    rabbitHost := getEnv("RABBIT_HOST", "rabbitmq")
    queueName := getEnv("RABBIT_QUEUE", "task_queue")
    rabbitURL := "amqp://guest:guest@" + rabbitHost + ":5672/"

    conn, err := amqp.Dial(rabbitURL)
    if err != nil {
        panic("Failed to connect to RabbitMQ: " + err.Error())
    }

    ch, err := conn.Channel()
    if err != nil {
        panic("Failed to open a channel: " + err.Error())
    }

    _, err = ch.QueueDeclare(
        queueName,
        true, false, false, false, nil,
    )
    if err != nil {
        panic("Failed to declare a queue: " + err.Error())
    }

    return ch, conn
}

func initMongoDB() *mongo.Client {
    mongoURI := getEnv("MONGO_URI", "mongodb://mongo:27017/")
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
    if err != nil {
        panic("Failed to connect to MongoDB: " + err.Error())
    }

    if err = client.Ping(context.Background(), nil); err != nil {
        panic("Failed to ping MongoDB: " + err.Error())
    }

    return client
}

func createTaskHandler(ch *amqp.Channel) gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... существующая логика обработчика POST /tasks
    }
}

func getTasksHandler(client *mongo.Client) gin.HandlerFunc {
    dbName := getEnv("DB_NAME", "mydb") // Получаем имя БД один раз
    
    return func(c *gin.Context) {
        activeRequests.Inc()
        defer activeRequests.Dec()

        timer := prometheus.NewTimer(httpDuration.WithLabelValues("/tasks"))
        defer timer.ObserveDuration()

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        collection := client.Database(dbName).Collection("tasks")
        // ... остальная логика обработчика GET /tasks
    }
}

