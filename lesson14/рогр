
---

# ✅ ЗАДАНИЕ: Добавьте Redis и реализуйте кэширование эндпоинтов `/notes` и `/tasks`

---

## 🎯 Цель

Улучшить наш сервис **Notes & Tasks**, добавив Redis для кэширования данных.

✅ Уменьшить нагрузку на PostgreSQL и MongoDB
✅ Ускорить ответы на популярные GET-запросы

---

## 💡 Описание

Мы хотим, чтобы при запросе:

* GET `/notes`
* GET `/tasks`

приложение сначала проверяло **Redis-кэш**:

* Если данные есть в кэше — вернуть их из Redis.
* Если нет — запросить из БД, сохранить результат в Redis с TTL и вернуть клиенту.

---

## 📦 Требования к функционалу

✅ Добавить Redis-сервис в `docker-compose.yml`
✅ Подключиться к Redis из Flask-приложения
✅ В GET-эндпоинтах `/notes` и `/tasks`:

* Проверять наличие кэша
* При попадании — отдавать кэш
* При промахе — идти в БД, записывать в Redis (с TTL = 60 сек), отдавать клиенту

✅ TTL для кэша = 60 секунд

---

## 🗂️ Пример структуры проекта (дополнено)

```
my_project/
│
├── docker-compose.yml
├── app/
│   ├── app.py
│   └── requirements.txt
├── nginx/
│   ├── nginx.conf
│   └── ssl/
│       ├── selfsigned.crt
│       └── selfsigned.key
└── README.md
```

---

## ⚙️ Переменные окружения

Добавьте в `.env`:

```
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_DB=0
```

---

## 🐳 Пример docker-compose.yml (дополненный)

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    ...

  mongo:
    image: mongo:6
    ...

  redis:
    image: redis:7
    ports:
      - "6379:6379"

  flask:
    build: ./app
    env_file:
      - .env
    ports:
      - "5001:5001"
    depends_on:
      - postgres
      - mongo
      - redis

  nginx:
    image: nginx:latest
    ...
```

---

## 🐍 Пример requirements.txt (дополненный)

```
Flask
psycopg2
pymongo
redis
python-dotenv
```

---

## 🐍 Пример кода (фрагмент Flask)

```python
import redis

# Redis
redis_client = redis.StrictRedis(
    host=os.getenv("REDIS_HOST"),
    port=int(os.getenv("REDIS_PORT")),
    db=int(os.getenv("REDIS_DB")),
    decode_responses=True
)
CACHE_TTL = 60
```

---

### 🔹 Пример для /notes (GET)

```python
@app.route('/notes', methods=['GET'])
def get_notes():
    cached = redis_client.get('notes_list')
    if cached:
        return cached, 200, {'Content-Type': 'application/json'}

    with get_pg_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("SELECT id, title, content, created_at FROM notes ORDER BY id")
            rows = cur.fetchall()

    notes = [
        {"id": r[0], "title": r[1], "content": r[2], "created_at": r[3].isoformat()}
        for r in rows
    ]

    import json
    redis_client.setex('notes_list', CACHE_TTL, json.dumps(notes))
    return jsonify(notes)
```

---

### 🔹 Пример для /tasks (GET)

```python
@app.route('/tasks', methods=['GET'])
def get_tasks():
    cached = redis_client.get('tasks_list')
    if cached:
        return cached, 200, {'Content-Type': 'application/json'}

    tasks = []
    for doc in tasks_collection.find():
        tasks.append({
            "id": str(doc["_id"]),
            "title": doc["title"],
            "description": doc["description"],
            "created_at": doc["created_at"].isoformat()
        })

    import json
    redis_client.setex('tasks_list', CACHE_TTL, json.dumps(tasks))
    return jsonify(tasks)
```

---

### 🔹 (Необязательно) Инвалидация кэша при POST

При создании новых заметок/задач — можно очистить кэш:

```python
@app.route('/notes', methods=['POST'])
def create_note():
    ...
    redis_client.delete('notes_list')
    return jsonify(...)

@app.route('/tasks', methods=['POST'])
def create_task():
    ...
    redis_client.delete('tasks_list')
    return jsonify(...)
```

---

## 📝 Чеклист для сдачи

✅ 1. Добавлен Redis в docker-compose
✅ 2. Подключение к Redis в Flask
✅ 3. Кэширование GET /notes и /tasks с TTL=60s
✅ 4. Инвалидация кэша при POST
✅ 5. README с инструкцией по запуску

---

## 🏁 Критерии приёмки

✅ Поднимается командой `docker-compose up`
✅ При повторных запросах GET /notes и /tasks данные берутся из Redis
✅ После POST запросов кэш сбрасывается

---

## 💬 Совет

> Сначала подними Redis в docker-compose и проверь подключение из Flask. Потом реализуй логику кэширования по шагам.

---
