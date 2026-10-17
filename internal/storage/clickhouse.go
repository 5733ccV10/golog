package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Config contains the settings for a native ClickHouse connection.
type Config struct {
	Address  string
	Database string
	Username string
	Password string
}

// Connect opens a pooled client and verifies connectivity before returning it.
// The caller owns the client and must close it after its users have stopped.
func Connect(ctx context.Context, cfg Config) (driver.Conn, error) {
	for _, setting := range []struct {
		name  string
		value string
	}{
		{"CLICKHOUSE_ADDR", cfg.Address},
		{"CLICKHOUSE_DB", cfg.Database},
		{"CLICKHOUSE_USER", cfg.Username},
		{"CLICKHOUSE_PASSWORD", cfg.Password},
	} {
		if setting.value == "" {
			return nil, fmt.Errorf("%s must be set", setting.name)
		}
	}

	const connectTimeout = 5 * time.Second
	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Protocol: clickhouse.Native,
		Addr:     []string{cfg.Address},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout: connectTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("open ClickHouse client: %w", err)
	}

	if err := conn.Ping(connectCtx); err != nil {
		pingErr := fmt.Errorf("ping ClickHouse: %w", err)
		if closeErr := conn.Close(); closeErr != nil {
			return nil, errors.Join(pingErr, fmt.Errorf("close ClickHouse client after failed ping: %w", closeErr))
		}
		return nil, pingErr
	}

	return conn, nil
}
