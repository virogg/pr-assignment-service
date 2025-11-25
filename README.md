# pr-assignment-service
Сервис назначения ревьюеров для Pull Request’ов. Микросервис, который автоматически назначает ревьюеров на Pull Request’ы (PR), а также позволяет управлять командами и участниками.

# Решение
- Реализована логика эндпоинтов из данного в задании API (OpenAPI-спецификация будет предоставлена отдельным файлом — `openapi.yaml`).
- Добавлены дополнительные эндпоинты статистики `/stats` и массовой деактивации пользователей команды `/team/deactivate`.
- База данных: PostgreSQL
- Реализованы миграции при помощи [goose](https://github.com/pressly/goose)
- Тесты:
  - Юнит ([uber-go/mock](https://github.com/uber-go/mock) для моков, `httptest` для тестирования HTTP-ручек)
  - Интеграциионные (с использованием `testcontainers-go`)
  - e2e
- Описана конфигурация линтера в `[.golangci.yml](.golangci.yml)`
- Реализован CI c помощью GitHub Actions, проверяющий линтер и все тесты

# Инструкция
Для запуска необходим `Docker` и `docker compose`
### Запуск
```bash
docker compose up --build
```
### Остановка
```bash
docker compose down
```

# Стек
- Go 1.25.4
- PostgreSQL
- goose
- Docker