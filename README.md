# DB Logger Tool

Простой и надежный логгер для Go приложений с автоматической фиксацией логов в PostgreSQL через `pgxpool`.

## Состав

В модуле:
type LogMsg struct {
	Module string - Имя модуля или приложения
	Level  string - Уровень лога(тип) ["INFO", "WARNING", "ERROR", "FATAL", "CRITICAL", итд]
	Code   int64  - Код сообщения
	Msg    string - Текст сообщения
	Debug  string - Отладочное поле (детали сообщения)
}

В БД:
	data_hash uuid NOT NULL, -- Хеш(При безусловной вставке - случайный),
		-- при вставке с перезаписью хеш от module, level, code, msg 
	occurence_count int8 DEFAULT 1 NOT NULL, - Количество повторений лога
	"module" text NOT NULL, -- Имя модуля или приложения
	"level" text NOT NULL,  --  Уровень лога(тип) ["INFO", "WARNING", "ERROR", "FATAL", "CRITICAL", итд]
	code int8 NOT NULL,  -- Код сообщения
	msg text NOT NULL, -- Текст сообщения
	debug text NULL, -- Отладочное поле перезаписываемое при обновлении (детали сообщения)
	upd_dttm timestamptz DEFAULT now() NOT NULL, -- Время вставки или обновления
	read_flg bool DEFAULT false NOT NULL,  -- Флаг о прочтении сообщения
	event_id uuid DEFAULT gen_random_uuid() NOT NULL, -- дополнительный UUD для гарантированного обновления флага о прочтении в 	-- ситуациях когда пользователь не прочитал еще сообщени, а лог уже обновиллся снова. 

## Логика вставки:
	event_id - Генерируется всегда (при вставке/обновления) служит для гарантированной пометки о чтении лога 
	upsert (true) - Если в БД уже есть строка где поля (module, level, code, msg) уже есть, происходит обновление счетчика,
		даты события, debug, event_id, read_flg - новая строка не вставляется
	upsert (false) - Вставка лога происходит в любом случае.
 
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
	myLogger.PushLog(ctx, msg, true)
	myLogger.PushLog(ctx, msg, false)
	// Будет вставленно 2 строки, т.к 2 схлопнуться в одну, 1 - вставиться в любом случае.
}
```