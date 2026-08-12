package logger

import (
	"io/fs"
	"embed"
	"fmt"
	"strings"
	"testing/fstest"

	// 1. Корневой пакет миграций
	"github.com/golang-migrate/migrate/v4"

	// 2. Драйвер iofs от golang-migrate (с подчеркиванием и без)
	_ "github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// 3. Драйвер pgx/v5 для миграций с алиасом
	migrate_pgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"

	// 4. Драйверы pgxpool и stdlib совместимости
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// MemoryFS подменяет текст файлов «на лету» для iofs драйвера
type MemoryFS struct {
	base        fs.FS
	targetTable string
	tableSuffix string
}

func (m *MemoryFS) Open(name string) (fs.File, error) {
	// Склеиваем путь для оригинального embed, так как там файлы лежат в папке migrations/
	originalPath := fmt.Sprintf("migrations/%s", name)
	
	file, err := m.base.Open(originalPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Если система пытается открыть директорию, возвращаем оригинальный объект
	if stat.IsDir() {
		return file, nil
	}

	buf := make([]byte, stat.Size())
	if _, err := file.Read(buf); err != nil {
		return nil, err
	}

	sqlStr := string(buf)
	sqlStr = strings.ReplaceAll(sqlStr, "{{TABLE_NAME}}", m.targetTable)
	sqlStr = strings.ReplaceAll(sqlStr, "{{TABLE_SUFFIX}}", m.tableSuffix)

	mockFS := fstest.MapFS{
		name: &fstest.MapFile{
			Data: []byte(sqlStr),
			Mode: stat.Mode(),
		},
	}

	return mockFS.Open(name)
}

// RunMigrations запускает миграции с динамической схемой
func RunMigrations(pool *pgxpool.Pool, targetTable string) error {
	tableSuffix := strings.ReplaceAll(targetTable, ".", "_")

	// 1. Создаем обертку для подмены строк в embed.FS
	memFS := &MemoryFS{
		base:        migrationFiles,
		targetTable: targetTable,
		tableSuffix: tableSuffix,
	}

	// 2. Инициализируем iofs драйвер от golang-migrate
	sourceDriver, err := iofs.New(memFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	// 3. Конвертируем пул в стандартный sql.DB
	sqlDB := stdlib.OpenDB(*pool.Config().ConnConfig)
	defer sqlDB.Close()

	// 4. Настраиваем pgx/v5 драйвер миграций через его алиас
	dbDriver, err := migrate_pgx.WithInstance(sqlDB, &migrate_pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create db driver: %w", err)
	}

	// 5. Инициализируем мигратор. Драйвер базы данных для pgx/v5 называется "pgx5"
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "pgx5", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	// 6. Накатываем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}