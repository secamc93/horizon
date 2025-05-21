package server

import (
	connectprostgres "cam/horizon/internal/infra/secundary/postgres/connectProstgres"
	"cam/horizon/pkg/env"
	"cam/horizon/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func RunServer() {
	log := logger.NewLogger()
	config := env.LoadEnv()
	port := config.ServerPort

	dbConn := setupDatabase()
	defer closeDatabase(dbConn, log)

	address := fmt.Sprintf(":%s", port)
	log.Info("Server running on port %s", port)

	coloredURL := formatSwaggerURL(address)
	log.Info("Documentacion swagger:\n\n%s\n", coloredURL)

	// setupSwagger(port)
	server := setupServer(address, log)
	startServer(server, log)
}

func formatSwaggerURL(address string) string {
	realURL := fmt.Sprintf("http://localhost%s/swagger/index.html", address)
	const width = 80
	leftPad := (width - len(realURL)) / 2
	centeredURL := fmt.Sprintf("%s%s", strings.Repeat(" ", leftPad), realURL)
	coloredURL := fmt.Sprintf("\033[32m%s\033[0m", centeredURL)
	return coloredURL
}

func setupDatabase() connectprostgres.DBConnection {
	dbConn := connectprostgres.New()
	return dbConn
}

func closeDatabase(dbConn connectprostgres.DBConnection, log logger.ILogger) {
	if err := dbConn.CloseDB(); err != nil {
		log.Error("Error closing DB: %v", err)
	}
}

// func setupSwagger(port string) {
// 	host := fmt.Sprintf("localhost:%s", port)
// 	// docs.SwaggerInfo.Host = host
// }

func setupServer(address string, log logger.ILogger) *http.Server {
	// gin.SetMode(gin.ReleaseMode)
	// gin.DefaultWriter = log.Writer()
	// gin.DefaultErrorWriter = log.Writer()
	// router := routers.SetupRouter()
	server := &http.Server{
		Addr: address,
		// Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	return server
}

func startServer(server *http.Server, log logger.ILogger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: %v", err)
	}

	log.Info("Server exiting")
}
