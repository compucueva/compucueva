package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "modernc.org/sqlite"

	"github.com/compucueva/compucueva/internal/api"
	"github.com/compucueva/compucueva/internal/db"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatalf("servidor finalizado con error: %v", err)
	}
}

func run(ctx context.Context) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "database.db"
	}

	// SQLITE connection
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("error al abrir SQLite: %w", err)
	}
	defer conn.Close()

	// SQLITE does not handle well multiple connections, do not change
	conn.SetMaxOpenConns(1)

	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("error al verificar SQLite: %w", err)
	}
	log.Printf("[db] Conectado exitosamente a %s", dbPath)

	queries := db.New(conn)
	apiServer := api.NewServer(queries)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: apiServer.Routes(),
	}

	// Start server
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("[api] Servidor escuchando en http://localhost:%s", port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for error
	select {
	case err := <-serverErr:
		return fmt.Errorf("fallo el servidor http: %w", err)
	case <-ctx.Done():
		log.Println("[api] Apagando el servidor...")
	}

	// Graceful shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("cierre forzado del servidor: %w", err)
	}

	log.Println("[api] Servidor detenido limpiamente")
	return nil
}
