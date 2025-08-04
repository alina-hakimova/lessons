from flask import Flask, request, jsonify, Response
import pika
import json
import os
import time
import logging
import sys
from pymongo import MongoClient
from datetime import datetime
from prometheus_client import Counter, Histogram, Gauge, make_wsgi_app, generate_latest
from werkzeug.middleware.dispatcher import DispatcherMiddleware

class JsonFormatter(logging.Formatter):
    def format(self, record):
        log_record = {
            "level": record.levelname,
            "time": self.formatTime(record, self.datefmt),
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "service": "flask-app"  # Добавляем имя сервиса для фильтрации
        }
        return json.dumps(log_record)

app = Flask(__name__)

# Настройка логирования
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(JsonFormatter())
app.logger.setLevel(logging.INFO)
app.logger.addHandler(handler)

# Создаем директорию для логов если ее нет
if not os.path.exists('logs'):
    os.makedirs('logs')

# Файловый логгер
file_handler = logging.FileHandler('logs/app.log')
file_handler.setFormatter(JsonFormatter())
app.logger.addHandler(file_handler)

# Prometheus metrics
REQUEST_COUNT = Counter(
    'flask_http_requests_total',
    'Total HTTP Requests',
    ['method', 'endpoint', 'status']
)

REQUEST_LATENCY = Histogram(
    'flask_http_request_duration_seconds',
    'HTTP request latency in seconds',
    ['endpoint']
)

ACTIVE_REQUESTS = Gauge(
    'flask_active_requests',
    'Active HTTP requests'
)

def init_rabbitmq(max_retries=5, retry_interval=5):
    """Инициализация подключения к RabbitMQ с повторными попытками"""
    retries = 0
    last_exception = None

    while retries < max_retries:
        try:
            connection = pika.BlockingConnection(
                pika.ConnectionParameters(
                    host=os.getenv("RABBIT_HOST"),
                    heartbeat=600,
                    blocked_connection_timeout=300
                )
            )
            channel = connection.channel()
            channel.queue_declare(
                queue=os.getenv("RABBIT_QUEUE"),
                durable=True
            )
            app.logger.info("Successfully connected to RabbitMQ")
            return connection, channel
        except pika.exceptions.AMQPConnectionError as e:
            last_exception = e
            retries += 1
            app.logger.error(f"RabbitMQ connection failed (attempt {retries}/{max_retries}): {str(e)}")
            if retries < max_retries:
                time.sleep(retry_interval)

    raise Exception(f"Failed to connect to RabbitMQ after {max_retries} attempts: {str(last_exception)}")

def init_mongodb(max_retries=5, retry_interval=5):
    """Инициализация подключения к MongoDB с повторными попытками"""
    retries = 0
    last_exception = None

    while retries < max_retries:
        try:
            client = MongoClient(
                os.getenv("MONGO_URI"),
                serverSelectionTimeoutMS=5000
            )
            client.admin.command('ping')
            db = client[os.getenv("DB_NAME")]
            app.logger.info("Successfully connected to MongoDB")
            return db
        except Exception as e:
            last_exception = e
            retries += 1
            app.logger.error(f"MongoDB connection failed (attempt {retries}/{max_retries}): {str(e)}")
            if retries < max_retries:
                time.sleep(retry_interval)

    raise Exception(f"Failed to connect to MongoDB after {max_retries} attempts: {str(last_exception)}")

# Инициализация подключений
try:
    rabbit_connection, rabbit_channel = init_rabbitmq()
    db = init_mongodb()
except Exception as e:
    app.logger.critical(f"Application startup failed: {str(e)}")
    exit(1)

@app.route('/tasks', methods=['POST'])
@REQUEST_LATENCY.labels(endpoint='/tasks').time()
def create_task():
    ACTIVE_REQUESTS.inc()
    try:
        data = request.get_json()
        app.logger.info(f"Received task creation request: {data}")

        if not data or 'title' not in data:
            REQUEST_COUNT.labels(method='POST', endpoint='/tasks', status='400').inc()
            app.logger.warning("Task creation failed - title is required")
            return jsonify({"error": "Title is required"}), 400

        task_data = {
            **data,
            "created_at": datetime.utcnow().isoformat(),
            "status": "queued"
        }

        rabbit_channel.basic_publish(
            exchange='',
            routing_key=os.getenv("RABBIT_QUEUE"),
            body=json.dumps(task_data),
            properties=pika.BasicProperties(
                delivery_mode=2,
                content_type='application/json'
            )
        )

        app.logger.info(f"Task queued successfully: {task_data}")
        REQUEST_COUNT.labels(method='POST', endpoint='/tasks', status='202').inc()
        return jsonify({"status": "queued", "task": task_data}), 202

    except pika.exceptions.AMQPChannelError as e:
        app.logger.error(f"RabbitMQ channel error: {str(e)}", exc_info=True)
        REQUEST_COUNT.labels(method='POST', endpoint='/tasks', status='500').inc()
        return jsonify({"error": f"RabbitMQ channel error: {str(e)}"}), 500
    except pika.exceptions.AMQPConnectionError as e:
        app.logger.error(f"RabbitMQ connection error: {str(e)}", exc_info=True)
        REQUEST_COUNT.labels(method='POST', endpoint='/tasks', status='503').inc()
        return jsonify({"error": f"RabbitMQ connection error: {str(e)}"}), 503
    except Exception as e:
        app.logger.error(f"Internal server error: {str(e)}", exc_info=True)
        REQUEST_COUNT.labels(method='POST', endpoint='/tasks', status='500').inc()
        return jsonify({"error": f"Internal server error: {str(e)}"}), 500
    finally:
        ACTIVE_REQUESTS.dec()

@app.route('/tasks', methods=['GET'])
@REQUEST_LATENCY.labels(endpoint='/tasks').time()
def get_tasks():
    ACTIVE_REQUESTS.inc()
    try:
        tasks = list(db.tasks.find({}, {'_id': 0}))
        app.logger.info(f"Retrieved {len(tasks)} tasks from database")
        REQUEST_COUNT.labels(method='GET', endpoint='/tasks', status='200').inc()
        return jsonify(tasks)
    except Exception as e:
        app.logger.error(f"Failed to fetch tasks: {str(e)}", exc_info=True)
        REQUEST_COUNT.labels(method='GET', endpoint='/tasks', status='500').inc()
        return jsonify({"error": f"Failed to fetch tasks: {str(e)}"}), 500
    finally:
        ACTIVE_REQUESTS.dec()

@app.route("/hello")
def hello():
    app.logger.info("Hello endpoint was called")
    return jsonify({"message": "Hello World"})

@app.route("/error")
def error():
    try:
        1 / 0
    except Exception as e:
        app.logger.error(f"Error occurred: {e}", exc_info=True)
        return jsonify({"error": str(e)}), 500

@app.route('/metrics')
def metrics():
    """Отдельный endpoint для метрик (альтернатива WSGI)"""
    return Response(
        generate_latest(),
        mimetype='text/plain; version=0.0.4'
    )

@app.teardown_appcontext
def close_connections(exception=None):
    if hasattr(app, 'rabbit_connection') and app.rabbit_connection.is_open:
        app.rabbit_connection.close()
    app.logger.info("Application shutdown - connections closed")

if __name__ == '__main__':
    app.rabbit_connection = rabbit_connection
    app.rabbit_channel = rabbit_channel
    app.db = db

    app.run(
        host='0.0.0.0',
        port=int(os.getenv("FLASK_PORT", 5001)),
        debug=os.getenv("FLASK_DEBUG", "false").lower() == "true"
    )
