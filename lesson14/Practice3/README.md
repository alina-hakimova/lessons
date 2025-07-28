
---

# ✅ ЗАДАНИЕ: Проверка ускорения с Redis и нагрузочное тестирование

---

## 🎯 Цель

Убедиться на практике, что Redis **ускоряет** ответы эндпоинтов `/notes` и `/tasks`, и измерить разницу во времени ответа.

✅ Увидеть разницу между обращением к БД и кэшем
✅ Проверить работу TTL в Redis
✅ Провести нагрузочное тестирование

---

## 💡 Описание

Сервис уже кэширует ответы в Redis. Сейчас мы:

✅ Добавим логирование времени запроса в Flask
✅ Посмотрим на реальные цифры (мс/µс)
✅ Проверим ключи и TTL в Redis через redis-cli
✅ Прогоним нагрузочные тесты через curl, Apache Bench (ab) или wrk

---

## 🗂️ Структура работы

✅ 1️⃣ Измерить время запроса в Flask
✅ 2️⃣ Проверить кэш через redis-cli
✅ 3️⃣ Запустить curl в цикле
✅ 4️⃣ Провести нагрузочное тестирование
✅ 5️⃣ Сравнить без кэша и с кэшем

---

## ⚙️ 1️⃣ Добавить измерение времени в Flask

В `app.py`:

```python
import time

@app.before_request
def start_timer():
    request.start_time = time.time()

@app.after_request
def log_request_time(response):
    if hasattr(request, 'start_time'):
        duration_ms = (time.time() - request.start_time) * 1000
        app.logger.info(f"{request.method} {request.path} took {duration_ms:.2f} ms")
    return response
```

✅ Результат: каждый запрос будет логироваться с временем обработки.

---

## ⚡️ 2️⃣ Проверка через curl

✅ Сделай серию запросов:

```bash
curl -k https://localhost/notes
```

✅ Посмотри логи приложения:

```
GET /notes took 25.30 ms
```

✅ Повтори запрос:

```
curl -k https://localhost/notes
```

✅ Лог покажет HIT из Redis:

```
GET /notes took 1.20 ms
```

---

## 💬 Объяснение

✅ Первый запрос → MISS → идёт в БД → дольше
✅ Повторный → HIT → берётся из Redis → быстрее

---

## 🔎 3️⃣ Проверка через redis-cli

✅ Зайди в контейнер:

```bash
docker-compose exec redis redis-cli
```

✅ Посмотри ключи:

```
keys *
```

✅ Пример:

```
1) "notes_list"
2) "tasks_list"
```

✅ Посмотри TTL:

```
ttl notes_list
```

✅ Пример вывода:

```
(integer) 55
```

✅ Проверяй, что TTL уменьшается и ключ исчезает через 60 секунд.

---

## 🗂️ 4️⃣ Проверка c удалением ключа

✅ Удали ключ:

```
del notes_list
```

✅ Сделай curl-запрос:

```
curl -k https://localhost/notes
```

✅ В логах:

```
Cache MISS
GET /notes took ~20–50 ms
```

✅ Повтори запрос:

```
curl -k https://localhost/notes
```

✅ В логах:

```
Cache HIT
GET /notes took ~1–2 ms
```

---

## 🚀 5️⃣ Цикл с curl

✅ Запусти:

```bash
for i in {1..5}; do curl -k https://localhost/notes; done
```

✅ Посмотри логи:

```
Cache HIT
Cache HIT
...
```

✅ Время ответа будет стабильно низким.

---

## ⚡️ 6️⃣ Нагрузочное тестирование

### 📌 Apache Bench (ab)

✅ Установи:

```bash
sudo apt install apache2-utils
```

✅ Пример теста:

```
ab -n 100 -c 10 https://localhost/notes
```

✅ В выводе:

```
Requests per second:    300.00 [#/sec]
Time per request:       3.33 [ms] (mean)
```

✅ Сравни при пустом кэше и при заполненном.

---

### 📌 wrk

✅ Установи:

```bash
sudo apt install wrk
```

✅ Пример:

```
wrk -t4 -c10 -d10s https://localhost/notes
```

✅ Вывод:

```
Latency   2.10ms ± 0.50ms
Requests/sec: 500.00
```

✅ Сравни:

* После очистки кэша (del notes\_list)
* После кэширования (несколько curl)

---

## 📌 7️⃣ Проверка через redis-cli во время теста

✅ Пока идёт ab или wrk:

```
docker-compose exec redis redis-cli
```

✅ Смотри TTL:

```
ttl notes_list
```

✅ Смотри количество ключей:

```
keys *
```

---

## ✅ Чеклист для сдачи

✅ Добавлено логирование времени в Flask
✅ Сделаны curl-запросы и сравнение времени
✅ Скриншоты или логи redis-cli с ключами и TTL
✅ Нагрузочные тесты (ab или wrk) — результат сравнения без кэша и с кэшем
✅ Вывод о разнице во времени ответа

---

## 🏁 Пример выводов (что написать в отчёте)

✅ Без кэша:

* GET /notes took \~30–50 ms
* RPS \~50–80

✅ С кэшем:

* GET /notes took \~1–2 ms
* RPS \~300–500

---

## 💬 Совет

> Для наглядности можно приложить скриншоты лога Flask, вывода redis-cli, результатов ab/wrk.

✅ Это докажет, что Redis реально ускоряет сервис!

---


