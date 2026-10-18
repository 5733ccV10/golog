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

	queue := make(chan storage.Entry, 20000)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           apis.NewRouter(queue, db),
		ReadTimeout:       29 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", server.Addr)

	ch := make(chan error, 1)

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- apis.ChannelWorker(workerCtx, db, queue)
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			ch <- fmt.Errorf("http server failed: %w", err)
		}
	}()

	workerStopped := false
	select {
	case err := <-ch:
		runErr = err
	case <-ctx.Done():
	case err := <-workerDone:
		workerStopped = true
		if err != nil {
			runErr = fmt.Errorf("worker unable to process requests: %w", err)
		} else {
			runErr = errors.New("log worker stopped unexpectedly")
		}
	}

	shutDownCtx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutDownCtx); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("HTTP shutdown failed: %w", err))
		if closeErr := server.Close(); closeErr != nil {
			runErr = errors.Join(runErr, fmt.Errorf("forced HTTP close failed: %w", closeErr))
		}
		cancelWorker()
	} else {
		log.Println("HTTP server stopped cleanly.")
		close(queue)
	}

	if workerStopped {
		return runErr
	}

	select {
	case err := <-workerDone:
		if err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("log worker stopped: %w", err))
		} else {
			log.Println("Log worker stopped cleanly.")
		}
	case <-shutDownCtx.Done():
		cancelWorker()
		workerErr := <-workerDone
		runErr = errors.Join(
			runErr,
			fmt.Errorf("worker shutdown timed out: %w", shutDownCtx.Err()),
			workerErr,
		)
	}
	return runErr
}
