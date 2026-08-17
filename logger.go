package logger

import (
	"context"
	"fmt"
	"log"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Logger struct {
	pool      *pgxpool.Pool
	fullName string
}

type Config struct {
	Schema    string // Схема в БД, например "public"
	TableName string // Имя таблицы, например "app_logs"
}

// NewLogger принимает пул и структуру конфигурации для БД
func NewLogger(pool *pgxpool.Pool, cfg Config) *Logger {
	if pool == nil {
		NewLogMsg("LOGGER", "FATAL", "pool is nil", -1, "NewLogger").Fatal()
	}

	// Выставляем дефолтные значения, если поля пустые
	if cfg.Schema == "" {
		cfg.Schema = "public"
	}
	if cfg.TableName == "" {
		cfg.TableName = "app_logs"
	}

	return &Logger{
		pool:     pool,
		fullName: fmt.Sprintf("%s.%s", cfg.Schema, cfg.TableName),
	}
}

/* PushLog - Вставка лога в БД
   Алгоритм вставки:
	event_id - Генерируется всегда (при вставке/обновления) служит для гарантированной пометки о чтении лога 
	upsert (true) - Если в БД уже есть строка где поля (module, level, code, msg) уже есть, происходит обновление счетчика,
		даты события, event_id, read_flg - новая строка не вставляется
	upsert (false) - Вставка лога происходит в любом случае. 
*/
func (l *Logger) PushLog(ctx context.Context, data *LogMsg, upsert bool) *LogMsg {
	var query string
	if upsert {
		query = `
			INSERT INTO ` + l.fullName + ` (data_hash, occurence_count, module, level, code, msg, debug, upd_dttm, read_flg, event_id)
			VALUES ($1, 1, $2, $3, $4, $5, $6, Now(), false, gen_random_uuid())
			ON CONFLICT (data_hash) DO UPDATE SET 
				occurence_count = ` + l.fullName + `.occurence_count + 1, 
				debug = EXCLUDED.debug, 
				upd_dttm = now(), 
				read_flg = false, 
				event_id = gen_random_uuid();`
	} else {
		query = `
			INSERT INTO ` + l.fullName + ` (data_hash, occurence_count, module, level, code, msg, debug, upd_dttm, read_flg, event_id)
			VALUES ($1, 1, $2, $3, $4, $5, $6, Now(), false, gen_random_uuid());`
	}

	_, err := l.pool.Exec(ctx, query, data.Hash(upsert), data.Module, data.Level, data.Code, data.Msg, data.Debug)
	if err != nil {
		log.Printf("error save log to db: %v\n", err)
		data.Print(true)
	}
	return data
}