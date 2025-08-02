from flask import Flask, request, jsonify, Response
import pika
import json
import os
import time
from pymongo import MongoClient
from datetime import datetime
from prometheus_client import Counter, Histogram, Gauge, make_wsgi_app, generate_latest
from werkzeug.middleware.dispatcher import DispatcherMiddleware

app = Flask(__name__)

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
            print("Successfully connected to RabbitMQ")
            return connection, channel
        except pika.exceptions.AMQPConnectionError as e:
            last_exception = e
            retries += 1
            print(f"RabbitMQ connection failed (attempt {retries}/{max_retries}): {str(e)}")
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
            print("Successfully connected to MongoDB")
            return db
        except Exception as e:
            last_exception = e
            retries += 1
            print(f"MongoDB connection failed (attempt {retries}/{max_retries}): {str(e)}")
            if retries < max_retries:
                time.sleep(retry_interval)

    raise Exception(f"Failed to connect to MongoDB after {max_retries} attempts: {str(last_exception)}")

# Инициализация подключений
try:
    rabbit_connection, rabbit_channel = init_rabbitmq()
    db = init_mongodb()
except Exception as e:
    print(f"Application startup failed: {str(e)}")
    exit(1)

@app.route('/tasks', methods=['POST'])
@REQUEST_LATENCY.labels(endpoint='/tasks').time()
def create_task():
    ACTIVE_REQUESTS.inc()
    try:
        data = request.get_json()

        if not data or 'title' not in data:
            REQUEST_COUNT.labels(method='POST', endpoint='/tasks', status='400').inc()
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

        REQUEST_COUNT.labels(method='POST', endpoint='/tasks', status='202').inc()
        return jsonify({"status": "queued", "task": task_data}), 202

    except pika.exceptions.AMQPChannelError as e:
        REQUEST_COUNT.labels(method='POST', endpoint='/tasks', status='500').inc()
        return jsonify({"error": f"RabbitMQ channel error: {str(e)}"}), 500
    except pika.exceptions.AMQPConnectionError as e:
        REQUEST_COUNT.labels(method='POST', endpoint='/tasks', status='503').inc()
        return jsonify({"error": f"RabbitMQ connection error: {str(e)}"}), 503
    except Exception as e:
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
        REQUEST_COUNT.labels(method='GET', endpoint='/tasks', status='200').inc()
        return jsonify(tasks)
    except Exception as e:
        REQUEST_COUNT.labels(method='GET', endpoint='/tasks', status='500').inc()
        return jsonify({"error": f"Failed to fetch tasks: {str(e)}"}), 500
    finally:
        ACTIVE_REQUESTS.dec()

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
    print("Application shutdown - connections closed")

if __name__ == '__main__':
    app.rabbit_connection = rabbit_connection
    app.rabbit_channel = rabbit_channel
    app.db = db

    app.run(
        host='0.0.0.0',
        port=int(os.getenv("FLASK_PORT", 5001)),
        debug=os.getenv("FLASK_DEBUG", "false").lower() == "true"
    )