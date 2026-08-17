# DB Logger Tool

Простой и надежный логгер для Go приложений с автоматической фиксацией логов в PostgreSQL через `pgxpool`.

## Состав

В модуле:  
```go
type LogMsg struct {
	Module string - Имя модуля или приложения 
	Level  string - Уровень лога(тип) ["INFO", "WARNING", "ERROR", "FATAL", "CRITICAL", итд] 
	Code   int64  - Код сообщения 
	Msg    string - Текст сообщения 
	Debug  string - Отладочное поле (детали сообщения) 
} 
```

В БД:
```sql  
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
	event_id uuid DEFAULT gen_random_uuid() NOT NULL, -- дополнительный UUD для гарантированного обновления флага о прочтении в 	-- ситуациях когда пользователь не прочитал еще сообщение, а лог уже обновился снова.  
```

## Логика вставки:
	event_id - Генерируется всегда (при вставке/обновления) служит для гарантированной пометки о чтении лога 
	upsert (true) - Если в БД уже есть строка где поля (module, level, code, msg) уже есть, происходит обновление счетчика,
		даты события, debug, event_id, read_flg - новая строка не вставляется
	upsert (false) - Вставка лога происходит в любом случае.
 
## 📌 Требования к БД (Важно)

Перед инициализацией логгера администратор базы данных должен **вручную создать схему и таблицу** и выдать права используемой учетной записи. 

SQL-команды:
```sql
CREATE TABLE IF NOT EXISTS {{TABLE_NAME}} (
	data_hash uuid NOT NULL,
	occurence_count int8 DEFAULT 1 NOT NULL,
	"module" text NOT NULL,
	"level" text NOT NULL,
	code int8 NOT NULL,
	msg text NOT NULL,
	debug text NULL,
	upd_dttm timestamptz DEFAULT now() NOT NULL,
	read_flg bool DEFAULT false NOT NULL,
	event_id uuid DEFAULT gen_random_uuid() NOT NULL,
	CONSTRAINT {{TABLE_SUFFIX}}_pkey PRIMARY KEY (data_hash)
);

CREATE INDEX IF NOT EXISTS idx_level_time_{{TABLE_SUFFIX}} ON {{TABLE_NAME}} USING btree (level, upd_dttm DESC);
CREATE INDEX IF NOT EXISTS idx_upd_dttm_{{TABLE_SUFFIX}} ON {{TABLE_NAME}} USING btree (upd_dttm DESC);

```

## 🚀 Установка

```bash
go get github.com/Rubashevskiy/logger
```

## 💻 Пример использования

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
