# DB Logger Tool

Простой и надежный логгер для Go приложений с автоматической фиксацией логов в PostgreSQL через `pgxpool`.

## 📌 Требования к БД (Важно)

Перед инициализацией логгера администратор базы данных должен **вручную создать схему** и выдать права используемой учетной записи. 

Пример SQL-команд:
```sql
-- 1. Создание схемы (Например test)
CREATE SCHEMA IF NOT EXISTS test;

-- 2. Выдача прав вашей учетной записи (например, app_user)
GRANT USAGE ON SCHEMA test TO app_user;
GRANT CREATE ON SCHEMA test TO app_user; -- Необходимо для работы golang-migrate (создание таблиц)
```

## 🚀 Установка

```bash
go get github.com/Rubashevskiy/logger
```

## 💻 Пример использования

Пакет автоматически создаст таблицу и необходимые индексы внутри указанной схемы при первом запуске (используется встроенный `golang-migrate`).

```go
package main

import (
	"context"
	"github.com/Rubashevskiy/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, _ := pgxpool.New(ctx, "postgres://app_user:pass@localhost:5432/mydb")

	// Инициализация логгера (таблица test.app_logs создастся сама)
	myLogger := logger.NewLogger(pool, logger.Config{
		Schema:    "test",
		TableName: "app_logs",
	})

	// Создание и отправка лога
	msg := logger.NewLogMsg("AUTH", "ERROR", "invalid password", 401, "ip: 127.0.0.1")
	myLogger.PushLog(ctx, msg, true)
}
```