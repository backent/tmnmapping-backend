package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/injector"
	servicesBuilding "github.com/malikabdulaziz/tmn-backend/services/building"
)

func main() {
	// Initialize logger FIRST - before any other operations
	helpers.InitLogger()
	logger := helpers.GetLogger()

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"warning": ".env file not found",
		}).Warn("Continuing without loading environment variables")
	}

	// Get application port
	APP_PORT := os.Getenv("APP_PORT")
	if APP_PORT == "" {
		APP_PORT = "8088"
	}

	// Get ERP sync interval
	syncIntervalStr := os.Getenv("ERP_SYNC_INTERVAL_MINUTES")
	syncInterval := 30
	if syncIntervalStr != "" {
		if val, err := strconv.Atoi(syncIntervalStr); err == nil {
			syncInterval = val
		}
	}

	// Initialize router with all dependencies
	router := injector.InitializeRouter()

	// Initialize building service for sync scheduler
	buildingService := injector.InitializeBuildingService()
	servicesBuilding.StartBuildingSyncScheduler(buildingService, helpers.Logger, syncInterval)

	// Create HTTP server
	server := http.Server{
		Addr:    ":" + APP_PORT,
		Handler: router,
	}

	// Start server
	logger.WithFields(map[string]interface{}{
		"port": APP_PORT,
	}).Info("Server is running")

	err = server.ListenAndServe()
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Server failed to start")
		panic(err)
	}
}
