package main

import (
    "context"
    "encoding/json"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/zsais/go-gin-prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/sirupsen/logrus"
    "github.com/streadway/amqp"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// initLogger инициализирует логгер
func initLogger() *logrus.Logger {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetOutput(os.Stdout)
    logger.SetLevel(logrus.InfoLevel)
    return logger
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

// loggingMiddleware добавляет логирование запросов
func loggingMiddleware() gin.HandlerFunc {
    logger := initLogger()
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        duration := time.Since(start)
        
        logger.WithFields(logrus.Fields{
            "method":   c.Request.Method,
            "path":     c.Request.URL.Path,
            "status":   c.Writer.Status(),
            "duration": duration,
        }).Info("Request processed")
    }
}

// initRabbitMQ инициализирует подключение к RabbitMQ
func initRabbitMQ() (*amqp.Channel, *amqp.Connection, error) {
    conn, err := amqp.Dial(getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"))
    if err != nil {
        return nil, nil, err
    }
    
    ch, err := conn.Channel()
    if err != nil {
        return nil, nil, err
    }
    
    return ch, conn, nil
}

// initMongoDB инициализирует подключение к MongoDB
func initMongoDB() (*mongo.Client, error) {
    clientOptions := options.Client().ApplyURI(getEnv("MONGODB_URL", "mongodb://mongodb:27017"))
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        return nil, err
    }
    
    err = client.Ping(context.Background(), nil)
    if err != nil {
        return nil, err
    }
    
    return client, nil
}

// createTaskHandler обрабатывает создание задачи
func createTaskHandler(ch *amqp.Channel) gin.HandlerFunc {
    return func(c *gin.Context) {
        var task map[string]interface{}
        if err := c.ShouldBindJSON(&task); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        body, err := json.Marshal(task)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        err = ch.Publish(
            "",                   // exchange
            "task_queue",         // routing key
            false,                // mandatory
            false,                // immediate
            amqp.Publishing{
                ContentType: "application/json",
                Body:        body,
            })
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(200, gin.H{"status": "Task created"})
    }
}

// getTasksHandler возвращает список задач
func getTasksHandler(client *mongo.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        collection := client.Database("tasks").Collection("tasks")
        cursor, err := collection.Find(context.Background(), bson.M{})
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        defer cursor.Close(context.Background())
        
        var tasks []bson.M
        if err = cursor.All(context.Background(), &tasks); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(200, tasks)
    }
}

func main() {
    logger := initLogger().WithFields(logrus.Fields{
        "service": "task-service",
        "version": "1.0.0",
    })
    logger.Info("Starting application")

    r := gin.Default()
    port := getEnv("GO_PORT", "5002")

    r.Use(loggingMiddleware())

    // Правильная инициализация Prometheus
    p := ginprometheus.NewPrometheus("gin")
    p.Use(r)

    ch, rabbitConn, err := initRabbitMQ()
    if err != nil {
        logger.WithError(err).Fatal("Failed to initialize RabbitMQ")
    }
    defer func() {
        if err := rabbitConn.Close(); err != nil {
            logger.WithError(err).Warn("Failed to close RabbitMQ connection")
        }
    }()
    defer func() {
        if err := ch.Close(); err != nil {
            logger.WithError(err).Warn("Failed to close RabbitMQ channel")
        }
    }()

    mongoClient, err := initMongoDB()
    if err != nil {
        logger.WithError(err).Fatal("Failed to initialize MongoDB")
    }
    defer func() {
        if err := mongoClient.Disconnect(context.Background()); err != nil {
            logger.WithError(err).Warn("Failed to disconnect MongoDB client")
        }
    }()

    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    r.POST("/tasks", createTaskHandler(ch))
    r.GET("/tasks", getTasksHandler(mongoClient))

    logger.WithField("port", port).Info("Server starting")
    if err := r.Run("0.0.0.0:" + port); err != nil {
        logger.WithError(err).Fatal("Server failed to start")
    }
}