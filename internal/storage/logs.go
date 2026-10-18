package storage

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

const insertLogsSQL = `INSERT INTO logs (service_id, occurred_at, received_at, message, level, metadata)`

func InsertBatch(ctx context.Context, conn driver.Conn, entries []Entry) error {
	batch, err := conn.PrepareBatch(ctx, insertLogsSQL)

	if err != nil {
		return err
	}
	defer batch.Close()

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
