package db

import (
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect baut die GORM-Verbindung zu PostgreSQL auf.
// Das DDL (Schema, Enum, Tabellen) wird über die Init-SQL-Skripte erzeugt,
// daher kein AutoMigrate hier (SQL-first).
func Connect(databaseURL string) (*gorm.DB, error) {
	database, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("verbindung zur datenbank fehlgeschlagen: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("zugriff auf datenbank-handle fehlgeschlagen: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("datenbank nicht erreichbar: %w", err)
	}

	slog.Info("Datenbankverbindung hergestellt")
	return database, nil
}

// Close schließt den zugrunde liegenden Verbindungspool.
func Close(database *gorm.DB) {
	if database == nil {
		return
	}
	if sqlDB, err := database.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
