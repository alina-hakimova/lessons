from flask import Flask, request, jsonify
import pika
import json
import os
import time
from pymongo import MongoClient
from datetime import datetime

app = Flask(__name__)

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
            # Проверяем подключение
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

# Инициализация подключений при старте приложения
try:
    rabbit_connection, rabbit_channel = init_rabbitmq()
    db = init_mongodb()
except Exception as e:
    print(f"Application startup failed: {str(e)}")
    exit(1)

@app.route('/tasks', methods=['POST'])
def create_task():
    try:
        data = request.get_json()
        
        if not data or 'title' not in data:
            return jsonify({"error": "Title is required"}), 400
        
        # Добавляем timestamp к задаче
        task_data = {
            **data,
            "created_at": datetime.utcnow().isoformat(),
            "status": "queued"
        }
        
        # Отправка задачи в очередь
        rabbit_channel.basic_publish(
            exchange='',
            routing_key=os.getenv("RABBIT_QUEUE"),
            body=json.dumps(task_data),
            properties=pika.BasicProperties(
                delivery_mode=2,  # Сообщение persistent
                content_type='application/json'
            )
        )
        
        return jsonify({
            "status": "queued",
            "task": task_data
        }), 202
    
    except pika.exceptions.AMQPChannelError as e:
        return jsonify({"error": f"RabbitMQ channel error: {str(e)}"}), 500
    except pika.exceptions.AMQPConnectionError as e:
        return jsonify({"error": f"RabbitMQ connection error: {str(e)}"}), 503
    except Exception as e:
        return jsonify({"error": f"Internal server error: {str(e)}"}), 500

@app.route('/tasks', methods=['GET'])
def get_tasks():
    try:
        tasks = list(db.tasks.find({}, {'_id': 0}))
        return jsonify(tasks)
    except Exception as e:
        return jsonify({"error": f"Failed to fetch tasks: {str(e)}"}), 500

@app.teardown_appcontext
def close_connections(exception=None):
    """Закрытие соединений при завершении работы приложения"""
    if hasattr(app, 'rabbit_connection') and app.rabbit_connection.is_open:
        app.rabbit_connection.close()
    print("Application shutdown - connections closed")

if __name__ == '__main__':
    # Сохраняем соединения в контексте приложения
    app.rabbit_connection = rabbit_connection
    app.rabbit_channel = rabbit_channel
    app.db = db
    
    app.run(
        host='0.0.0.0',
        port=int(os.getenv("FLASK_PORT", 5001)),
        debug=os.getenv("FLASK_DEBUG", "false").lower() == "true"
    )