

# 📊 Мониторинг микросервисов: 4 Golden Signals с Prometheus и Grafana


## 🎯 Цель

Освоить внедрение мониторинга 4 Golden Signals для микросервисов, научиться:

- Внедрять экспорт метрик в Python и Go сервисах.
- Собрать метрики Prometheus.
- Визуализировать основные сигналы в Grafana-дэшборде.


## 📦 Исходные данные

- **Python-сервис**: работает на Flask, получает данные из PostgreSQL.
- **Go-сервис**: работает на net/http, получает данные из MongoDB.


## 🧱 Требования к реализации

### 1. Добавь метрики в сервисы

Для каждого сервиса требуется:

- Экспонировать endpoint `/metrics` (формат Prometheus).
- Использовать официальный Prometheus-клиент:
    - [`prometheus_client`](https://github.com/prometheus/client_python) для Python
    - [`prometheus/client_golang`](https://github.com/prometheus/client_golang) для Go
- Реализовать метрики для 4 Golden Signals:

| Signal | Что измеряем | Тип метрики |
| :-- | :-- | :-- |
| Latency | Латентность HTTP-запроса | Histogram/Summary |
| Traffic | Количество HTTP-запросов (RPS) | Counter |
| Errors | Кол-во запросов с 4xx/5xx | Counter |
| Saturation | Загруженность (например, задачи) | Gauge |

### 2. Примеры кода метрик

**Python:**

```python
from flask import Flask, jsonify, request
import random, time
from prometheus_client import Counter, Summary, Gauge, generate_latest

app = Flask(__name__)

REQUEST_COUNT = Counter('http_requests_total', 'Total HTTP requests', ['method', 'endpoint', 'status'])
REQUEST_LATENCY = Summary('http_request_latency_seconds', 'Latency of HTTP requests', ['endpoint'])
ACTIVE_TASKS = Gauge('active_tasks_count', 'Current active tasks')

@app.route('/task', methods=['POST'])
@REQUEST_LATENCY.labels(endpoint='/task').time()
def create_task():
    ACTIVE_TASKS.inc()
    try:
        time.sleep(random.uniform(0.1, 0.4))
        if random.random() < 0.2:
            REQUEST_COUNT.labels(method='POST', endpoint='/task', status='500').inc()
            return 'Error', 500
        REQUEST_COUNT.labels(method='POST', endpoint='/task', status='201').inc()
        return 'OK', 201
    finally:
        ACTIVE_TASKS.dec()

@app.route('/metrics')
def metrics():
    return generate_latest(), 200, {'Content-Type': 'text/plain; version=0.0.4'}
```

**Go:**

```go
var (
    requests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "go_http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    latency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "go_http_request_latency_seconds",
            Help:    "Request latency",
        },
        []string{"endpoint"},
    )
)
func dataHandler(w http.ResponseWriter, r *http.Request) {
    timer := prometheus.NewTimer(latency.WithLabelValues("/data"))
    defer timer.ObserveDuration()
    // Your business logic...
    if rand.Float64() < 0.1 {
        requests.WithLabelValues("GET", "/data", "500").Inc()
        http.Error(w, "internal error", 500)
        return
    }
    requests.WithLabelValues("GET", "/data", "200").Inc()
    w.Write([]byte("OK"))
}
```


## 🐳 3. Docker Compose

Создай структуру проекта:

```
monitoring-task/
├── docker-compose.yml
├── prometheus/
│   └── prometheus.yml
├── python_service/
│   └── app.py
├── go_service/
│   └── main.go
```

**docker-compose.yml:**

```yaml
version: '3.8'

services:
  python-service:
    build: ./python_service
    ports: ["8000:8000"]

  go-service:
    build: ./go_service
    ports: ["8080:8080"]

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
    ports: ["9090:9090"]

  grafana:
    image: grafana/grafana:latest
    ports: ["3000:3000"]
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
```


## 🔧 4. Конфиг Prometheus

**prometheus/prometheus.yml:**

```yaml
global:
  scrape_interval: 5s

scrape_configs:
  - job_name: 'python-service'
    static_configs:
      - targets: ['python-service:8000']

  - job_name: 'go-service'
    static_configs:
      - targets: ['go-service:8080']
```


## 📈 5. Дэшборд Grafana: 4 Golden Signals

В Grafana создай новый дэшборд с 4 основными графиками для каждого сервиса:

### 1. 🚀 Traffic (RPS)

**PromQL:**

```
sum by (method) (rate(http_requests_total[1m]))
```


### 2. ⏱️ Latency (Latency P90/P99)

**PromQL:**

```
histogram_quantile(0.9, sum by (le) (rate(http_request_latency_seconds_bucket[1m])))
```


### 3. ❌ Errors (Error count/rate)

**PromQL:**

```
sum by (status) (rate(http_requests_total{status=~"4..|5.."}[1m]))
```


### 4. 📉 Saturation (Current load)

**PromQL:**

```
active_tasks_count
```

## 🧪 6. Как проверить?

1. Перейди на http://localhost:8000/task и http://localhost:8080/data для проверки сервисов.
2. Зайди на http://localhost:9090 и убедись, что Prometheus видит оба сервиса и их метрики `/metrics`.
3. Открой http://localhost:3000 (Grafana, логин/пароль: admin/admin), подключи Prometheus и построй дашборд из вышеуказанных графиков.



