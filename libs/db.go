package libs

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/malikabdulaziz/tmn-backend/helpers"
)

func NewDatabase() *sql.DB {
	logger := helpers.GetLogger()

	MYSQL_HOST := os.Getenv("MYSQL_HOST")
	MYSQL_PORT := os.Getenv("MYSQL_PORT")
	MYSQL_USER := os.Getenv("MYSQL_USER")
	MYSQL_PASSWORD := os.Getenv("MYSQL_PASSWORD")
	MYSQL_DATABASE := os.Getenv("MYSQL_DATABASE")

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", MYSQL_USER, MYSQL_PASSWORD, MYSQL_HOST, MYSQL_PORT, MYSQL_DATABASE)

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"host":     MYSQL_HOST,
			"port":     MYSQL_PORT,
			"database": MYSQL_DATABASE,
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
			"host":     MYSQL_HOST,
			"port":     MYSQL_PORT,
			"database": MYSQL_DATABASE,
			"error":    err.Error(),
		}).Error("Failed to ping database")
		panic(err)
	}

	logger.WithFields(map[string]interface{}{
		"host":                  MYSQL_HOST,
		"port":                  MYSQL_PORT,
		"database":              MYSQL_DATABASE,
		"max_open_connections":  DB_MAX_OPEN_CONNECTIONS,
		"max_idle_connections":  DB_MAX_IDLE_CONNECTIONS,
		"conn_max_lifetime_sec": DB_CONN_MAX_LIFETIME_IN_SEC,
	}).Info("Database connection established successfully")

	return db
}

