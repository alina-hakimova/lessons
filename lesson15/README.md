# ✅ ЗАДАНИЕ: Добавьте RabbitMQ и реализуйте очередь задач

## 🌟 Цель

Познакомиться с брокером сообщений RabbitMQ и использовать его для асинхронной обработки задач в нашем сервисе Notes & Tasks.

* ✅ Научиться публиковать сообщения в очередь
* ✅ Запустить отдельный воркер, который их обрабатывает
* ✅ Разделить приложение на producer (Flask) и consumer (worker)

---

## 💡 Описание

Мы хотим, чтобы при создании новой задачи (POST /tasks) она сначала попадала в очередь RabbitMQ, а отдельный фоновый воркер ёё забирал и сохранял в MongoDB.

---

## 🧰 Микросервисная архитектура

В этом уроке мы впервые добавляем новый микросервис — worker.py. Это:

* отдельный компонент системы,
* создаёт обработку задач в фоне,
* позволяет разгрузить Flask от посторонних операций.

### Producer & Consumer

* Flask — это producer. Он отвечает на HTTP запрос и отправляет сообщение в RabbitMQ.
* worker.py — это consumer. Он слушает RabbitMQ и обрабатывает задачи.

---

## 📆 Структура проекта

my_project/
├── docker-compose.yml
├── .env
├── app/
│   ├── app.py          # Flask producer
│   ├── worker.py       # Consumer — отдельный микросервис
│   └── requirements.txt
├── nginx/
│   └── ...
└── README.md

---

## ⚙️ .env

RABBIT_HOST=rabbitmq
RABBIT_PORT=5672
RABBIT_QUEUE=task_queue

---

## 📦 docker-compose.yml (фрагмент)

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      - RABBITMQ_DEFAULT_USER=guest
      - RABBITMQ_DEFAULT_PASS=guest

  flask:
    build: ./app
    env_file:
      - .env
    ports:
      - "5001:5001"
    depends_on:
      - rabbitmq
      - mongo

  worker:
    build: ./app
    command: python worker.py
    env_file:
      - .env
    depends_on:
      - rabbitmq
      - mongo

---

## 📜 requirements.txt (фрагмент)

Flask
pika
pymongo
python-dotenv

---

## 🐉 Flask (producer)

import pika, json, os

rabbit_connection = pika.BlockingConnection(
    pika.ConnectionParameters(host=os.getenv("RABBIT_HOST"))
)
channel = rabbit_connection.channel()
channel.queue_declare(queue=os.getenv("RABBIT_QUEUE"), durable=True)

@app.route('/tasks', methods=['POST'])
def create_task():
    data = request.get_json()
    channel.basic_publish(
        exchange='',
        routing_key=os.getenv("RABBIT_QUEUE"),
        body=json.dumps(data),
        properties=pika.BasicProperties(delivery_mode=2)
    )
    return jsonify({"status": "queued"}), 202

---

## 🛠 worker.py (consumer)

import pika, os, json
from pymongo import MongoClient
from datetime import datetime

mongo_client = MongoClient("mongodb://mongo:27017/")
db = mongo_client["mydb"]
tasks_collection = db["tasks"]

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host=os.getenv("RABBIT_HOST"))
)
channel = connection.channel()
channel.queue_declare(queue=os.getenv("RABBIT_QUEUE"), durable=True)

def callback(ch, method, properties, body):
    task = json.loads(body)
    task["created_at"] = datetime.utcnow()
    tasks_collection.insert_one(task)
    print(f"[x] Task saved: {task['title']}")
    ch.basic_ack(delivery_tag=method.delivery_tag)

channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue=os.getenv("RABBIT_QUEUE"), on_message_callback=callback)

print(" [*] Waiting for messages. To exit press CTRL+C")
channel.start_consuming()

---

## 📊 Чеклист

* [x] RabbitMQ добавлен в docker-compose
* [x] Flask публикует задачи в RabbitMQ
* [x] worker.py подключен и сохраняет задачи в MongoDB
* [x] Очередь durable
* [x] README описывает запуск воркера

---

## 🌼 Совет

Если не используете worker как сервис в compose, запустите ручно:

docker exec -it <flask-container> python app/worker.py

UI для проверки RabbitMQ: [http://localhost:15672](http://localhost:15672) (guest / guest)
