import pika
import os
import json
from pymongo import MongoClient
from datetime import datetime
import time

def connect_to_rabbitmq():
    """Устанавливает соединение с RabbitMQ с повторными попытками"""
    attempts = 0
    max_attempts = 5
    wait_time = 5
    
    while attempts < max_attempts:
        try:
            connection = pika.BlockingConnection(
                pika.ConnectionParameters(host=os.getenv("RABBIT_HOST"))
            )
            return connection
        except pika.exceptions.AMQPConnectionError:
            attempts += 1
            print(f"Connection failed, attempt {attempts}/{max_attempts}. Retrying in {wait_time} seconds...")
            time.sleep(wait_time)
    
    raise Exception("Could not connect to RabbitMQ")

def connect_to_mongodb():
    """Устанавливает соединение с MongoDB"""
    client = MongoClient(os.getenv("MONGO_URI"))
    db = client[os.getenv("DB_NAME")]
    return db.tasks

def process_task(ch, method, properties, body):
    """Обрабатывает полученную задачу"""
    try:
        task = json.loads(body)
        print(f" [x] Received task: {task}")
        
        # Добавляем метаданные
        task["processed_at"] = datetime.utcnow()
        task["status"] = "completed"
        
        # Сохраняем в MongoDB
        tasks_collection.insert_one(task)
        print(f" [x] Task saved to MongoDB: {task['title']}")
        
        # Подтверждаем обработку
        ch.basic_ack(delivery_tag=method.delivery_tag)
    except Exception as e:
        print(f" [x] Error processing task: {e}")

if __name__ == '__main__':
    # Подключение к MongoDB
    tasks_collection = connect_to_mongodb()
    
    # Подключение к RabbitMQ
    connection = connect_to_rabbitmq()
    channel = connection.channel()
    
    # Объявление очереди (на случай если worker запустится раньше producer)
    channel.queue_declare(queue=os.getenv("RABBIT_QUEUE"), durable=True)
    
    # Настройка QoS
    channel.basic_qos(prefetch_count=1)
    
    # Подписка на очередь
    channel.basic_consume(
        queue=os.getenv("RABBIT_QUEUE"),
        on_message_callback=process_task
    )
    
    print(' [*] Waiting for messages. To exit press CTRL+C')
    channel.start_consuming()
