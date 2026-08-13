# DB Logger Tool

Простой и надежный логгер для Go приложений с автоматической фиксацией логов в PostgreSQL через `pgxpool`.

## 📌 Требования к БД (Важно)

Перед инициализацией логгера администратор базы данных должен **вручную создать схему** и выдать права используемой учетной записи. 

Пример SQL-команд:
```sql
--  Разрешаем пользователю заходить в схему public
GRANT USAGE ON SCHEMA public TO app_user;

--  Разрешаем создавать в ней новые объекты (чтобы создалась таблица schema_migrations)
GRANT CREATE ON SCHEMA public TO app_user;
--  Разрешаем пользователю заходить в схему public
GRANT USAGE ON SCHEMA test_log TO app_user;

-- Даем право создавать в ней таблицы (включая динамические)
GRANT CREATE ON SCHEMA test_log TO app_user;

-- На случай, если таблица app_logs или связанные с ней индексы/последовательности уже существуют:
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA test_log TO app_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA test_log TO app_user;

-- Права на все таблицы, которые будут созданы в test_log в будущем этим или другими миграциями
ALTER DEFAULT PRIVILEGES IN SCHEMA test_log GRANT ALL PRIVILEGES ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA test_log GRANT ALL PRIVILEGES ON SEQUENCES TO app_user;
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