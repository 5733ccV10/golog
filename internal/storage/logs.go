package storage

import (
	"context"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

const insertLogsSQL = `
INSERT INTO logs (service_id, occurred_at, received_at, message, level, metadata)
`

func InsertBatch(ctx context.Context, conn driver.Conn, entries []Entry) error {
	batch, err := conn.PrepareBatch(ctx, insertLogsSQL)

	if err != nil {
		return err
	}
	defer func() {
		if err := batch.Close(); err != nil {
			log.Printf("close query results: %v", err)
		}
	}()

	for _, entry := range entries {
		err := batch.Append(
			entry.ServiceID,
			entry.OccurredAt,
			entry.ReceivedAt,
			entry.Message,
			string(entry.Level),
			entry.Metadata,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

type LogFilter struct {
	ServiceID string
	Level     Level
	From      time.Time
	Till      time.Time
	Count     int
}

const fetchFromLogs = `
SELECT service_id, occurred_at, message, level, metadata
FROM logs
WHERE service_id = ?
  AND occurred_at >= ?
  AND occurred_at < ?
`

func GetFilteredLogs(ctx context.Context, conn driver.Conn, filter LogFilter) ([]Entry, error) {
	query := fetchFromLogs
	args := []any{filter.ServiceID, filter.From, filter.Till}

	if filter.Level != "" {
		query += " AND level = ?"
		args = append(args, string(filter.Level))
	}

	query += " ORDER BY occurred_at DESC LIMIT ?"
	args = append(args, filter.Count)
	rows, err := conn.Query(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close query results: %v", err)
		}
	}()

	entries := make([]Entry, 0, filter.Count)

	for rows.Next() {
		var entry Entry
		var level string

		err = rows.Scan(
			&entry.ServiceID,
			&entry.OccurredAt,
			&entry.Message,
			&level,
			&entry.Metadata,
		)
		if err != nil {
			return nil, err
		}
		entry.Level = Level(level)
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
