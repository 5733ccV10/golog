package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/5733ccV10/go-log/internal/apis"
	"github.com/5733ccV10/go-log/internal/storage"
)

func runServer() (runErr error) {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := storage.Connect(ctx, storage.Config{
		Address:  os.Getenv("CLICKHOUSE_ADDR"),
		Database: os.Getenv("CLICKHOUSE_DB"),
		Username: os.Getenv("CLICKHOUSE_USER"),
		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
	})
	if err != nil {
		return fmt.Errorf("initialize storage: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("close ClickHouse client: %w", err))
		}
	}()
	log.Println("Connected to ClickHouse.")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           apis.NewRouter(),
		ReadTimeout:       29 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", server.Addr)

	ch := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			ch <- fmt.Errorf("http server failed: %w", err)
		}
	}()

	select {
	case err := <-ch:
		return err

	case <-ctx.Done():
		shutDownCtx, cancel := context.WithTimeout(
			context.Background(), 5*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutDownCtx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				log.Printf("Forced server close failed: %v", closeErr)
			}
			return fmt.Errorf("Server forced to shutdown with error: %w", err)
		}

		log.Println("HTTP server stopped cleanly.")
	}
	return nil
}
