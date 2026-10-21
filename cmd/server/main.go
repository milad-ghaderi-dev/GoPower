package main

import (
	"GoPower/db"
	"GoPower/repository"
	"GoPower/server"
	"GoPower/service"
	"GoPower/ws"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Structured Logging Setup (Standard library log/slog)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting GoPower Smart Grid Service...")

	// 2. Read Configuration from Environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "gopower.db"
	}

	// 3. Database Initialization
	database, err := db.InitDB(dbPath)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	slog.Info("Database initialized successfully", "path", dbPath)

	// 4. Initialize Repositories and Domain Services
	stRepo := repository.NewStationRepository(database)
	csRepo := repository.NewConsumerRepository(database)
	gridService := service.NewGridService(database, stRepo, csRepo)

	// 5. Start Real-time WebSocket Hub
	hub := ws.NewWebSocketHub()
	go hub.Run()
	slog.Info("WebSocket hub started")

	// 6. Router & HTTP Server Setup
	router := server.SetupRouter(gridService, stRepo, csRepo, hub)
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Run Server in Background Goroutine
	go func() {
		slog.Info("HTTP server listening", "port", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// 8. Graceful Shutdown on OS Signals (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("Shutdown signal received", "signal", sig.String())

	// Stop accepting new HTTP requests with 5-second deadline
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server graceful shutdown failed", "error", err)
	} else {
		slog.Info("HTTP server gracefully stopped")
	}

	// Stop WebSocket hub and close all client connections
	hub.Stop()
	slog.Info("WebSocket hub stopped")

	slog.Info("GoPower service exited cleanly")
}
