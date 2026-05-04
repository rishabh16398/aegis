package db

import (
	"fmt"

	"github.com/aegis-av/aegis/internal/db/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	conn *gorm.DB
}

func New(path string) (*DB, error) {
	conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite at %s: %w", path, err)
	}

	// WAL mode: allows concurrent reads while a write is happening.
	// Foreign keys: enforce referential integrity (off by default in SQLite).
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA foreign_keys=ON")

	if err := migrate(conn); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Conn returns the underlying *gorm.DB so other packages can run queries.
func (d *DB) Conn() *gorm.DB {
	return d.conn
}

func (d *DB) Close() error {
	sql, err := d.conn.DB()
	if err != nil {
		return err
	}
	return sql.Close()
}

// migrate runs AutoMigrate for every model.
// AutoMigrate creates the table if it doesn't exist, and adds any new
// columns — but it never drops columns or changes existing ones.
func migrate(conn *gorm.DB) error {
	return conn.AutoMigrate(
		&models.ScanJob{},
		&models.ScanResult{},
		&models.Threat{},
		&models.QuarantineEntry{},
		&models.NetworkEvent{},
		&models.BlockedIP{},
		&models.ProcessEvent{},
		&models.Alert{},
	)
}
