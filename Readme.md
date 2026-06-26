# Запуск

```bash
cd deploy

export SUBSERV_DB_USER=...
export SUBSERV_DB_PASSWORD=...
export SUBSERV_DB_NAME=...
export SUBSERV_HTTP_PORT=...
export SUBSERB_DB_URL=...

# Или в .env

# Опционально использовать config.yaml с установленной переменной SUBSERV_CONFIG_PATH=config.yaml

docker compose up -d
```

# Особенности

- Delete отсутствующей подписки - не ошибка
- Автогенерация спецификации OpenAPI `go run ./cmd/genspec/main.go`
- Multi stage dockerfile
- Ошибки внутренних слоёв не скрыты
- Не обрабатывается проблема lost update при конкурентных вызовах update одной подписки
- List с курсором вместо offset + limit
- Конфигурация из .yaml и переменных окружения
- Глобальный логгер вместо передачи логера в сервис, репозиторий


# Конфигурация

```yaml
db:
    url: postgres://...
    maxconnections: 50 # Опционально

http:
    host: localhost # Опционально
    port: 33220 # Опционально
    docs: false # Опционально - web ui для API допступно по /docs

app:
    log: "info" # / "warn" / "debug"
```

также через env

```bash
SUBSERV_DB_URL=...
SUBSERV_DB_MAXCONNECTIONS=...
...
```
