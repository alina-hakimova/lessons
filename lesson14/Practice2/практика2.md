

---

# ✅ 1️⃣ Примеры полных функций (Flask)

---

## Импорт, настройки Redis в app.py

```python
import redis
import json

# Redis клиент
redis_client = redis.StrictRedis(
    host=os.getenv("REDIS_HOST"),
    port=int(os.getenv("REDIS_PORT")),
    db=int(os.getenv("REDIS_DB")),
    decode_responses=True
)
CACHE_TTL = 60  # секунды
```

---

## ✅ GET /notes

```python
@app.route('/notes', methods=['GET'])
def get_notes():
    # 1️⃣ Попытка достать из кэша
    cached_data = redis_client.get('notes_list')
    if cached_data:
        app.logger.info("Cache HIT for /notes")
        return cached_data, 200, {'Content-Type': 'application/json'}

    # 2️⃣ Если нет — идем в Postgres
    with get_pg_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("SELECT id, title, content, created_at FROM notes ORDER BY id")
            rows = cur.fetchall()

    notes = [
        {"id": r[0], "title": r[1], "content": r[2], "created_at": r[3].isoformat()}
        for r in rows
    ]

    # 3️⃣ Кэшируем результат
    redis_client.setex('notes_list', CACHE_TTL, json.dumps(notes))
    app.logger.info("Cache MISS for /notes. Set new cache.")
    return jsonify(notes)
```

---

## ✅ POST /notes (с инвалидированием кэша)

```python
@app.route('/notes', methods=['POST'])
def create_note():
    data = request.get_json()
    with get_pg_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                "INSERT INTO notes (title, content) VALUES (%s, %s) RETURNING id, created_at",
                (data['title'], data['content'])
            )
            row = cur.fetchone()
    # Удаляем кэш
    redis_client.delete('notes_list')
    app.logger.info("Cache invalidated for /notes after POST")
    return jsonify({"id": row[0], "title": data['title'], "content": data['content'], "created_at": row[1].isoformat()})
```

---

## ✅ GET /tasks

```python
@app.route('/tasks', methods=['GET'])
def get_tasks():
    cached_data = redis_client.get('tasks_list')
    if cached_data:
        app.logger.info("Cache HIT for /tasks")
        return cached_data, 200, {'Content-Type': 'application/json'}

    tasks = []
    for doc in tasks_collection.find():
        tasks.append({
            "id": str(doc["_id"]),
            "title": doc["title"],
            "description": doc["description"],
            "created_at": doc["created_at"].isoformat()
        })

    redis_client.setex('tasks_list', CACHE_TTL, json.dumps(tasks))
    app.logger.info("Cache MISS for /tasks. Set new cache.")
    return jsonify(tasks)
```

---

## ✅ POST /tasks (с инвалидированием кэша)

```python
@app.route('/tasks', methods=['POST'])
def create_task():
    data = request.get_json()
    task = {
        "title": data['title'],
        "description": data['description'],
        "created_at": datetime.utcnow()
    }
    result = tasks_collection.insert_one(task)
    task['_id'] = str(result.inserted_id)
    
    redis_client.delete('tasks_list')
    app.logger.info("Cache invalidated for /tasks after POST")
    return jsonify(task)
```

---

# ✅ 2️⃣ Пример готового README.md

---

```markdown
# Notes & Tasks Microservice

Flask-сервис для управления заметками (PostgreSQL) и задачами (MongoDB) с кэшированием в Redis.

---

## 🚀 Что умеет

✅ POST /notes — сохраняет заметку в Postgres  
✅ GET /notes — возвращает список заметок (с кэшем в Redis)  
✅ POST /tasks — сохраняет задачу в MongoDB  
✅ GET /tasks — возвращает список задач (с кэшем в Redis)  
✅ Прокси через nginx с HTTPS (self-signed)

---

## 🐳 Быстрый старт

1️⃣ Клонируй репозиторий  
2️⃣ Добавь файл `.env`:

```

POSTGRES\_HOST=postgres
POSTGRES\_DB=notes\_db
POSTGRES\_USER=notes\_user
POSTGRES\_PASSWORD=mystrongpassword

MONGO\_URI=mongodb://mongo\_user\:mongopass\@mongo:27017/
MONGO\_DB=tasks\_db

REDIS\_HOST=redis
REDIS\_PORT=6379
REDIS\_DB=0

````

3️⃣ Собери и запусти:

```bash
docker-compose up --build
````

---

## 📦 Сервисы в docker-compose

✅ PostgreSQL
✅ MongoDB
✅ Redis
✅ Flask-приложение
✅ nginx с SSL

---

## 🌐 Пример HTTPS-запросов

✅ Создать заметку:

```bash
curl -k -X POST https://localhost/notes -H "Content-Type: application/json" -d '{"title":"Test Note","content":"Hello Postgres"}'
```

✅ Получить все заметки (кэшируется):

```bash
curl -k https://localhost/notes
```

✅ Создать задачу:

```bash
curl -k -X POST https://localhost/tasks -H "Content-Type: application/json" -d '{"title":"Buy milk","description":"From MongoDB"}'
```

✅ Получить все задачи (кэшируется):

```bash
curl -k https://localhost/tasks
```

---

## 💾 Как работает кэш

✅ При первом GET-запросе:

* Flask берёт из БД
* Сохраняет результат в Redis на 60 секунд
* Возвращает клиенту

✅ При повторных GET-запросах:

* Отдаёт из Redis за миллисекунды

✅ При POST-запросах:

* Очищает кэш, чтобы следующие GET вернули актуальные данные

---

## 📝 Пример docker-compose.yml

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

## ✅ Критерии приёмки

✅ Поднимается через `docker-compose up`
✅ Работает через nginx по HTTPS
✅ /notes и /tasks поддерживают кэширование в Redis с TTL=60s
✅ POST-запросы инвалидируют кэш
✅ Документация (этот файл)

---

````

---

# ✅ 3️⃣ Подсказки по тестированию с redis-cli

---

После запуска:

```bash
docker-compose exec redis redis-cli
````

---

### 📌 Посмотреть все ключи

```
keys *
```

✅ Должны появляться:

```
"notes_list"
"tasks_list"
```

---

### 📌 Посмотреть содержимое кэша

```
get notes_list
get tasks_list
```

✅ Увидишь JSON.

---

### 📌 Проверить TTL

```
ttl notes_list
ttl tasks_list
```

✅ Должно показывать время жизни (<= 60 секунд).

---

### 📌 Удалить ключ вручную

```
del notes_list
del tasks_list
```

✅ Для ручного сброса кэша.

---

### 📌 Проверить, что POST инвалидирует

1️⃣ GET → кешируется
2️⃣ redis-cli → keys \* ✅ ключ есть
3️⃣ POST → создаёт запись
4️⃣ redis-cli → keys \* ✅ ключа нет

---

## 🚀 Совет для менти

> Используй redis-cli в контейнере, чтобы наглядно отлаживать кэш. Проверь работу TTL и автоудаление.

---

Если хочешь, можем сделать ещё:

✅ Полный docker-compose файл
✅ Полный пример app.py
✅ Инструкции для деплоя на сервер


