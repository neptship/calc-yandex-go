# сalc-yandex-go

Cервис на Go для вычисления арифметических выражений. Сервис принимает математические выражения через POST-запросы и возвращает вычисленные результаты.

## Функциональность

- Асинхронное вычисление арифметических выражений
- Поддержка базовых арифметических операций (+, -, *, /)
- Поддержка скобок для управления порядком операций
- Формат обмена данными JSON
- Распределённое выполнение операций между агентами
- История вычислений с подробной информацией
- Веб-интерфейс для удобного использования
- Обработка ошибок с соответствующими HTTP-кодами

## Установка и запуск

### Предварительные требования

- Установленный Go (версия 1.20 или выше)
- Node.js 18+ и npm (для фронтенда)
- Docker и Docker Compose (для запуска через Docker)
- Make (опционально, для упрощения запуска)

### Запуск через Make
Для удобства запуска всех компонентов системы используйте команды Make:

```bash
# Клонирование репозитория
git clone https://github.com/neptship/calc-yandex-go
cd calc-yandex-go

# Установка всех зависимостей (Go модулей и npm пакетов)
make install

# Запуск всех компонентов (оркестратор, агент, фронтенд)
make run-all
```

### Запуск через Docker
Проект поддерживает запуск через Docker и Docker Compose:

```bash
git clone https://github.com/neptship/calc-yandex-go
cd calc-yandex-go
docker-compose up
```

### Ручной запуск

**Клонирование репозитория и установка зависимостей**

```bash
# Установка Go модулей
go mod tidy

# Установка npm пакетов
cd frontend
npm install
cd ..
```

**Запуск оркестратора**
```bash
go run cmd/orchestrator/main.go
```

**Запуск агентов (В отдельном терминале)**

```bash
go run cmd/agent/main.go
```

По умолчанию оркестратор запускается на порту 8080.

### Запуск фронтенда

```bash
cd frontend
npm run dev
```

Веб-интерфейс будет доступен по адресу: http://localhost:3000

## API Спецификация

```mermaid
flowchart LR
    A[Фронтенд Next.js] -->|Отправка выражения| B[Оркестратор Go]
    B -->|Задачи с указанным временем| C[Агенты Go]
    C -->|Результаты операций| B
    B -->|Итоговый ответ| A
```
### POST /api/v1/calculate

Отправляет выражение на вычисление и возвращает идентификатор задачи.

**Формат запроса:**
```json
{
    "expression": "2+2*2"
}
```

**Успешный ответ (201 Created):**
```json
{
    "id": 1
}
```

**Ответ при некорректном выражении (422 Unprocessable Entity):**
```json
{
    "error": "invalid expression"
}
```

**Ответ при внутренней ошибке сервера (500 Internal Server Error):**
```json
{
    "error": "internal server error"
}
```

### GET /api/v1/expressions/:id

Получает статус и результат вычисления по идентификатору.

**Успешный ответ (200 OK), вычисление завершено:**

```json
{
    "expression": {
        "id": 1,
        "status": "completed",
        "result": 6
    }
}
```

**Успешный ответ (200 OK), вычисление в процессе:**

```json
{
    "expression": {
        "id": 1,
        "status": "processing"
    }
}
```

**Выражение не найдено (404 Not Found):**

```json
{
    "error": "Expression not found"
}
```

### GET /api/v1/expressions

Получает список всех выражений и их статусов.

**Успешный ответ (200 OK):**

```json
{
    "expressions": [
        {
            "id": 1,
            "status": "completed",
            "result": 6
        },
        {
            "id": 2,
            "status": "processing"
        }
    ]
}
```

### GET /internal/task
Получает задачу для выполнения агентом.

**Успешный ответ (200 OK):**
```json
{
    "task": {
        "id": 1,
        "arg1": 2,
        "arg2": 2,
        "operation": "+",
        "operation_time": 1000
    }
}
```

**Ответ, если нет доступных задач (404 Not Found):**
```json
{
    "error": "No tasks available"
}
```

### POST /internal/task
Отправляет результат выполнения задачи.

**Формат запроса:**
```json
{
    "id": 1,
    "result": 4
}
```
**Успешный ответ (200 OK):**
```json
{
    "success": true
}
```
**Ответ, если задача не найдена (404 Not Found):**
```json
{
    "error": "Task not found"
}
```


## Примеры использования

### Успешные сценарии

**Отправка выражения на вычисление**
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
    "expression": "2+2*2"
}'
```

Ответ:
```json
{
    "id": 1
}
```

**Отправка сложного выражения**
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
    "expression": "(70/7) * 10 /((3+2) * (3+7)) - 2"
}'
```

Ответ:
```json
{
    "id": 2
}
```

**Получение результата вычисления**
```bash
curl --location 'http://localhost:8080/api/v1/expressions/1'
```

Ответ:
```json
{
    "expression": {
        "id": 1,
        "status": "completed",
        "result": 6
    }
}
```

**Получение списка всех выражений**
```bash
curl --location 'http://localhost:8080/api/v1/expressions'
```

Ответ:
```json
{
    "expressions": [
        {
            "id": 1,
            "status": "completed",
            "result": 6
        },
        {
            "id": 2,
            "status": "processing"
        }
    ]
}
```

**Получение задачи агентом**
```bash
curl --location 'http://localhost:8080/internal/task'
```

Ответ:
```json
{
    "task": {
        "id": 3,
        "arg1": 3,
        "arg2": 5,
        "operation": "+",
        "operation_time": 1000
    }
}
```

**Отправка результата задачи**
```bash
curl --location 'http://localhost:8080/internal/task' \
--header 'Content-Type: application/json' \
--data '{
    "id": 3,
    "result": 8
}'
```

Ответ:
```json
{
    "success": true
}
```

### Обработка ошибок

**Невалидный JSON при отправке выражения**
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data 'not-valid-json'
```

Ответ:
```json
{
    "error": "Invalid request format"
}
```

### Некорректное выражение
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
    "expression": "2+*2"
}'
```

Ответ:
```json
{
    "error": "invalid expression"
}
```

**Запрос несуществующего выражения**
```bash
curl --location 'http://localhost:8080/api/v1/expressions/999'
```

Ответ:
```json
{
    "error": "Expression not found"
}
```

**Деление на ноль**
```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
    "expression": "5/0"
}'
```

После проверки статуса выражения:
```json
{
    "expression": {
        "id": 3,
        "status": "failed",
        "result": null
    }
}
```

**Запрос задачи, когда нет доступных задач**
```bash
curl --location 'http://localhost:8080/internal/task'
```

Ответ:
```json
{
    "error": "No tasks available"
}
```

**Некорректный запрос при отправке результата задачи**
```bash
curl --location 'http://localhost:8080/internal/task' \
--header 'Content-Type: application/json' \
--data 'not-valid-json'
```

Ответ:
```json
{
    "error": "Invalid request body"
}
```

**Отправка результата для несуществующей задачи**
```bash
curl --location 'http://localhost:8080/internal/task' \
--header 'Content-Type: application/json' \
--data '{
    "id": 999,
    "result": 42
}'
```

Ответ:
```json
{
    "error": "Task not found"
}
```

## Настройка

Система настраивается через переменные окружения:
- `PORT` - порт для HTTP-сервера (по умолчанию 8080)
- `COMPUTING_POWER` - количество параллельных вычислителей в агенте (по умолчанию 3)
- `TIME_ADDITION_MS` - время выполнения сложения (по умолчанию 1000 мс)
- `TIME_SUBTRACTION_MS` - время выполнения вычитания (по умолчанию 1000 мс)
- `TIME_MULTIPLICATIONS_MS` - время выполнения умножения (по умолчанию 1500 мс)
- `TIME_DIVISIONS_MS` - время выполнения деления (по умолчанию 2000 мс)


## Ограничения

- Поддерживаются только положительные целые числа
- Использование унарного минуса или плюса приведет к некорректной работе
- Поддерживаются только POST-запросы
- Все нестандартные символы в выражении (буквы, спецсимволы) приведут к ошибке 422

## Веб-интерфейс
Веб-интерфейс предоставляет следующие возможности:
- Ввод арифметических выражений
- Отображение прогресса вычислений в реальном времени
- История вычислений с результатами
- Просмотр подробной информации в JSON-формате
- Очистка истории

## Запуск тестов

```bash
# Запуск всех тестов
cd calc-yandex-go
make test

# или напрямую
go test ./... -v
```

## Рекомендации по тестированию

- Рекомендуется использовать Postman для тестирования API, так как с curl могут возникнуть проблемы
- При использовании curl рекомендуется выполнять запросы через git bash терминал
- Для тестирования асинхронных вычислений сделайте запрос, а затем периодически запрашивайте результат

## Структура проекта
```
/calc-yandex-go/
├── cmd/                     # Точки входа для исполняемых файлов
│   ├── orchestrator/        # Оркестратор
│   └── agent/               # Агент
├── internal/                # Внутренние пакеты
│   ├── agent/               # Логика агента
│   ├── config/              # Конфигурация
│   ├── models/              # Модели данных
│   └── orchestrator/        # Логика оркестратора
├── pkg/                     # Переиспользуемые пакеты
│   └── calculation/         # Парсинг и вычисление выражений
├── frontend/                # Next.js фронтенд
└── docker-compose.yml       # Конфигурация Docker
```