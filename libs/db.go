package libs

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/malikabdulaziz/tmn-backend/helpers"
)

func NewDatabase() *sql.DB {
	logger := helpers.GetLogger()

	POSTGRES_HOST := os.Getenv("POSTGRES_HOST")
	POSTGRES_PORT := os.Getenv("POSTGRES_PORT")
	POSTGRES_USER := os.Getenv("POSTGRES_USER")
	POSTGRES_PASSWORD := os.Getenv("POSTGRES_PASSWORD")
	POSTGRES_DATABASE := os.Getenv("POSTGRES_DATABASE")
	POSTGRES_SSLMODE := os.Getenv("POSTGRES_SSLMODE")

	dataSourceName := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_HOST, POSTGRES_PORT, POSTGRES_DATABASE, POSTGRES_SSLMODE)

	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"host":     POSTGRES_HOST,
			"port":     POSTGRES_PORT,
			"database": POSTGRES_DATABASE,
			"error":    err.Error(),
		}).Error("Failed to open database connection")
		panic(err)
	}

	// Get connection pool settings from environment
	DB_CONN_MAX_LIFETIME_IN_SEC, err := strconv.Atoi(os.Getenv("DB_CONN_MAX_LIFETIME_IN_SEC"))
	if err != nil {
		DB_CONN_MAX_LIFETIME_IN_SEC = 300 // Default 5 minutes
	}

	DB_MAX_OPEN_CONNECTIONS, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNECTIONS"))
	if err != nil {
		DB_MAX_OPEN_CONNECTIONS = 10 // Default
	}

	DB_MAX_IDLE_CONNECTIONS, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNECTIONS"))
	if err != nil {
		DB_MAX_IDLE_CONNECTIONS = 5 // Default
	}

	// Connection pool settings
	db.SetConnMaxLifetime(time.Second * time.Duration(DB_CONN_MAX_LIFETIME_IN_SEC))
	db.SetMaxOpenConns(DB_MAX_OPEN_CONNECTIONS)
	db.SetMaxIdleConns(DB_MAX_IDLE_CONNECTIONS)

	// Test connection
	err = db.Ping()
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"host":     POSTGRES_HOST,
			"port":     POSTGRES_PORT,
			"database": POSTGRES_DATABASE,
			"error":    err.Error(),
		}).Error("Failed to ping database")
		panic(err)
	}

	logger.WithFields(map[string]interface{}{
		"host":                  POSTGRES_HOST,
		"port":                  POSTGRES_PORT,
		"database":              POSTGRES_DATABASE,
		"max_open_connections":  DB_MAX_OPEN_CONNECTIONS,
		"max_idle_connections":  DB_MAX_IDLE_CONNECTIONS,
		"conn_max_lifetime_sec": DB_CONN_MAX_LIFETIME_IN_SEC,
	}).Info("Database connection established successfully")

	return db
}
