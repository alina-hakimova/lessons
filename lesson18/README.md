# 📘 Задание: Настроить логирование веб-приложения через Grafana Loki

---

## 🎯 Цель

Научиться собирать, хранить и просматривать логи из веб-приложения на Flask через связку **Fluent Bit + Loki + Grafana**.

---

## 🔧 Часть 1: Подготовка docker-compose окружения

### 📁 Структура проекта

```
monitoring/
├── docker-compose.yml
├── grafana/
│   └── provisioning/
│       └── datasources/
│           └── loki.yaml
└── fluent-bit/
    ├── fluent-bit.conf
    └── parsers.conf

app/
└── app.py  # Flask-приложение
```

---

## 🚀 Часть 2: Flask-приложение с JSON логами

### 📄 `app/app.py`

```python
from flask import Flask, jsonify, request
import logging
import sys
import json

class JsonFormatter(logging.Formatter):
    def format(self, record):
        log_record = {
            "level": record.levelname,
            "time": self.formatTime(record, self.datefmt),
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName
        }
        return json.dumps(log_record)

app = Flask(__name__)

handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(JsonFormatter())
app.logger.setLevel(logging.INFO)
app.logger.addHandler(handler)

@app.route("/hello")
def hello():
    app.logger.info("Hello endpoint was called")
    return jsonify({"message": "Hello World"})

@app.route("/error")
def error():
    try:
        1 / 0
    except Exception as e:
        app.logger.error(f"Error occurred: {e}")
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
```

---

## ⚙️ Часть 3: Конфигурация Fluent Bit

### 📄 `monitoring/fluent-bit/fluent-bit.conf`

```ini
[SERVICE]
    Flush        1
    Daemon       Off
    Log_Level    info
    Parsers_File parsers.conf

[INPUT]
    Name              tail
    Path              /var/log/app/*.log
    Parser            json
    Tag               flask.*

[FILTER]
    Name              kubernetes
    Match             *
    Merge_Log         On

[OUTPUT]
    Name              loki
    Match             *
    Url               http://loki:3100/loki/api/v1/push
    BatchWait         1
    LabelKeys         level,module,function
```

### 📄 `monitoring/fluent-bit/parsers.conf`

```ini
[PARSER]
    Name        json
    Format      json
    Time_Key    time
    Time_Format %Y-%m-%dT%H:%M:%S
```

---

## 📦 Часть 4: Grafana и Loki

### 📄 `monitoring/grafana/provisioning/datasources/loki.yaml`

```yaml
apiVersion: 1

datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    isDefault: true
```

---

## 🐳 Часть 5: docker-compose.yml

### 📄 `monitoring/docker-compose.yml`

```yaml
version: '3.8'

services:
  loki:
    image: grafana/loki:2.9.0
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml

  fluent-bit:
    image: fluent/fluent-bit:2.2
    volumes:
      - ./fluent-bit/fluent-bit.conf:/fluent-bit/etc/fluent-bit.conf
      - ./fluent-bit/parsers.conf:/fluent-bit/etc/parsers.conf
      - ../app/logs:/var/log/app

  grafana:
    image: grafana/grafana:10.2.0
    ports:
      - "3000:3000"
    volumes:
      - ./grafana/provisioning:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin

  flask-app:
    build:
      context: ../app
    volumes:
      - ../app/logs:/app/logs
    command: python app.py
```

---

## 🧱 Часть 6: Dockerfile для Flask

### 📄 `app/Dockerfile`

```dockerfile
FROM python:3.10

WORKDIR /app

COPY app.py .

RUN pip install flask

CMD ["python", "app.py"]
```

---

## 🧪 Часть 7: Как проверить

1. Запусти:

```bash
cd monitoring
docker-compose up --build
```

2. Открой Grafana: [http://localhost:3000](http://localhost:3000)
   - Логин: `admin`
   - Пароль: `admin`

3. Перейди в **Explore**, выбери Loki и напиши:

```
{module="app"}
```

4. Проверь приложение:

```bash
curl http://localhost:5000/hello
curl http://localhost:5000/error
```

---

## 📊 Что проверить ментору

- [ ] Приложение пишет JSON-логи
- [ ] Fluent Bit видит логи и отправляет в Loki
- [ ] Grafana отображает логи в Explore
- [ ] Присутствуют лейблы: `level`, `function`, `module`

---

## 🧠 Дополнительно

- Логи в JSON удобно парсить и фильтровать
- Loki использует лейблы как ключи фильтрации (аналог меток Prometheus)
- Можно писать алерты на основе логов (например, много `ERROR` → сигнал тревоги)

