
# Multi-Language Web App with HAProxy Load Balancer

Проект разворачивает 4 веб-сервера на разных языках (Python, Node.js, Go, Java) с балансировкой нагрузки через HAProxy. Все серверы отображают секретное сообщение из переменных окружения.

## 📋 Требования
- Ubuntu/Debian сервер (в примере используется `192.168.116.131`)
- Ansible 2.9+
- Docker
- Доступ по SSH с ключом

## 🛠️ Структура проекта
```

.
├── group_vars/
│   ├── all.yml       \# Основные переменные
│   └── vault.yml     \# Секретные данные
├── inventory.ini     \# Инвентарь серверов
├── playbook.yml      \# Главный плейбук
└── roles/
├── python_app/   \# Python-сервер (Flask)
├── node_app/     \# Node.js-сервер
├── go_app/       \# Go-сервер
├── java_app/     \# Java-сервер
└── haproxy/      \# Настройка балансировщика

```

## 🚀 Быстрый старт
1. Клонируй репозиторий:
   ```bash
   git clone <repo-url> && cd nginx-configs/lesson8
```

2. Настрой инвентарь (`inventory.ini`):

```ini
[staging_servers]
linux1 ansible_host=192.168.116.131 ansible_user=alina ansible_ssh_private_key_file=/home/alina2/.ssh/alina2
```

3. Запусти плейбук:

```bash
ansible-playbook -i inventory.ini playbook.yml
```


## 🔍 Проверка работы

1. Тест балансировки (выполни несколько раз):

```bash
curl http://192.168.116.131
```

Пример ответа:

```text
Secret message from Go: Hello, secure world!
```

2. Открой в браузере:

```
http://192.168.116.131/haproxy?stats
```

![HAProxy Stats](https://i.imgur.com/JZ7m3lD.png)
3. Проверь все контейнеры:

```bash
docker ps
```

Должны быть видны 4 сервера:

```text
python_app (8001)
node_app   (8002)
go_app     (8003)
java_app   (8004)
```


## ⚙️ Основные настройки

Все параметры редактируются в `group_vars/all.yml`:

```yaml
secret_message: "Hello, secure world!"
app_ports:
  python: 8001
  node: 8002
  go: 8003
  java: 8004
haproxy_frontend_port: 80
domain_name: helloworld.local
```


## 🛠️ Ручное управление

- Перезагрузить HAProxy:

```bash
sudo systemctl reload haproxy
```

- Остановить все серверы:

```bash
docker stop python_app node_app go_app java_app
```

- Логи HAProxy:

```bash
tail -f /var/log/haproxy.log
```


## 💡 Особенности проекта

- Все серверы запускаются в Docker-контейнерах
- Автоматическая сборка образов через Ansible
- Round Robin балансировка нагрузки
- Мониторинг через HAProxy Stats (порт 1936)


## ⚠️ Устранение проблем

Если сервис не работает:

1. Проверь конфиг HAProxy:

```bash
sudo haproxy -c -f /etc/haproxy/haproxy.cfg
```

2. Проверь логи контейнеров:

```bash
docker logs python_app  # и другие имена контейнеров
```

3. Убедись, что порты не заняты:

```bash
ss -tulnp | grep '80\|800'
```

Можешь сразу сохранить его в корень проекта (`~/nginx-configs/lesson8/README.md`) и при необходимости адаптировать под свои нужды
