package logger

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// Регистрируем и импортируем стандартный iofs драйвер
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs"

	// Драйвер pgx/v5 с алиасом для устранения конфликта имен
	migrate_pgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var rawMigrationFiles embed.FS

// CustomFileInfo реализует интерфейс fs.FileInfo для наших модифицированных файлов
type CustomFileInfo struct {
	name string
	size int64
	mode fs.FileMode
}

func (c *CustomFileInfo) Name() string       { return c.name }
func (c *CustomFileInfo) Size() int64        { return c.size }
func (c *CustomFileInfo) Mode() fs.FileMode  { return c.mode }
func (c *CustomFileInfo) ModTime() time.Time { return time.Now() }
func (c *CustomFileInfo) IsDir() bool        { return false }
func (c *CustomFileInfo) Sys() any          { return nil }

// MemoryFile реализует интерфейс fs.File для чтения SQL из памяти
type MemoryFile struct {
	reader *bytes.Reader
	info   fs.FileInfo
}

func (m *MemoryFile) Stat() (fs.FileInfo, error) { return m.info, nil }
func (m *MemoryFile) Read(b []byte) (int, error) { return m.reader.Read(b) }
func (m *MemoryFile) Close() error               { return nil }

// CustomFileSystem перехватывает чтение файлов и подменяет плейсхолдеры
type CustomFileSystem struct {
	baseFS      fs.FS
	targetTable string
	tableSuffix string
}

func (c *CustomFileSystem) Open(name string) (fs.File, error) {
	file, err := c.baseFS.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return file, nil
	}

	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Динамически заменяем плейсхолдеры схемы и таблицы
	sqlStr := string(buf)
	sqlStr = strings.ReplaceAll(sqlStr, "{{TABLE_NAME}}", c.targetTable)
	sqlStr = strings.ReplaceAll(sqlStr, "{{TABLE_SUFFIX}}", c.tableSuffix)

	// Возвращаем кастомный файл из памяти
	return &MemoryFile{
		reader: bytes.NewReader([]byte(sqlStr)),
		info: &CustomFileInfo{
			name: stat.Name(),
			size: int64(len(sqlStr)),
			mode: stat.Mode(),
		},
	}, nil
}

// RunMigrations выполняет миграции
func RunMigrations(pool *pgxpool.Pool, targetTable string) error {
	// Изолируем папку migrations, чтобы убрать префиксы путей
	subFS, err := fs.Sub(rawMigrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to get sub migrations dir: %w", err)
	}

	tableSuffix := strings.ReplaceAll(targetTable, ".", "_")

	customFS := &CustomFileSystem{
		baseFS:      subFS,
		targetTable: targetTable,
		tableSuffix: tableSuffix,
	}

	// Передаем точку ".", так как файлы уже находятся в корне благодаря fs.Sub
	sourceDriver, err := iofs.New(customFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	sqlDB := stdlib.OpenDB(*pool.Config().ConnConfig)
	defer sqlDB.Close()

	dbDriver, err := migrate_pgx.WithInstance(sqlDB, &migrate_pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "pgx5", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}