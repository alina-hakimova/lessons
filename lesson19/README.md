
Этот проект представляет собой мультисервисное приложение с двумя сервисами: на Python (Flask) и Go. Каждый сервис упакован в отдельный Docker-образ, сборка и тестирование которых автоматизированы с помощью CI/CD в GitLab (файл `.gitlab-ci.yml`). Для запуска используется `docker-compose`.

***

## 📁 Структура проекта

```
my-project/
├── .gitlab-ci.yml         # CI/CD пайплайн для сборки, тестирования и деплоя
├── .gitignore             # Исключения для git (кеши, временные файлы, логи и т.п.)
├── docker-compose.yml     # Описание сервисов и сетей Docker
├── python/                # Папка с Python-сервисом
│   ├── Dockerfile         # Dockerfile для Python приложения
│   ├── requirements.txt   # Зависимости Python
│   ├── main.py            # Основной код Python сервиса (Flask)
│   └── test_main.py       # Тесты для Python сервиса
├── go/                    # Папка с Go-сервисом
│   ├── Dockerfile         # Dockerfile для Go приложения
│   ├── main.go            # Основной код Go сервиса
│   ├── main_test.go       # Тесты для Go сервиса
│   └── go.mod             # Модуль Go
```


***

## 🔧 CI/CD `.gitlab-ci.yml`

Пайплайн состоит из трех этапов:

- **build** — сборка докер-образов для Python и Go сервисов.
- **test** — запуск юнит-тестов для обоих сервисов.
- **deploy** — остановка старых контейнеров и запуск новых с помощью `docker-compose`.

***

## 🐍 Python сервис

- Использует Flask для веб-сервера.
- Запускается на порту 8000.
- Реализованы два эндпоинта:
    - `/` — возвращает "Hello from Python Service!"
    - `/health` — возвращает JSON `{"status": "healthy"}`.
- В папке `python` есть тесты на эти энпоинты.

***

## ⚙️ Go сервис

- Веб-сервер с двумя аналогичными эндпоинтами на порту 8080:
    - `/` — возвращает "Hello from Go Service!"
    - `/health` — возвращает JSON `{"status": "healthy"}`.
- Есть тесты в `go/main_test.go`.

***

## 🐳 Docker Compose

Определены два сервиса — `python-app` и `go-app` с пробросом портов:

- Python сервис доступен на порту 8000.
- Go сервис — на 8080.

Образы используют теги на основе хеша коммита `$CI_COMMIT_SHA`.

***

## Как запустить проект

1. Клонируйте репозиторий и перейдите в ветку lesson19:
```bash
git clone https://github.com/alina-hakimova/lessons.git
cd lessons
git checkout lesson19
```

2. Запустите сборку и запуск контейнеров через Docker Compose:
```bash
docker-compose up -d --build
```

3. Проверьте доступность сервисов:

- Python: http://localhost:8000
- Go: http://localhost:8080

***
Все скрипты и конфигурации находятся в репозитории по ссылке:
https://github.com/alina-hakimova/lessons.git (ветка lesson19)
opped completely. Someone may choose to fork your project or volunteer to step in as a maintainer or owner, allowing your project to keep going. You can also make an explicit request for maintainers.
