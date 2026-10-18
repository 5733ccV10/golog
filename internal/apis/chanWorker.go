package apis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/5733ccV10/go-log/internal/storage"
)

const batchSize int = 10000

func ChannelWorker(ctx context.Context, conn driver.Conn, queue <-chan storage.Entry) error {
	batch := make([]storage.Entry, 0, batchSize)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		started := time.Now()
		if err := storage.InsertBatch(ctx, conn, batch); err != nil {
			return fmt.Errorf("insert batch: %w", err)
		}
		log.Printf("batch=%d insert=%s queued=%d",
			len(batch), time.Since(started), len(queue))
		batch = batch[:0]
		ticker.Reset(500 * time.Millisecond)
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case entry, ok := <-queue:
			if ok == false {
				if err := flush(); err != nil {
					return fmt.Errorf("failed while inserting final batch: %w", err)
				}

				return nil
			}
			batch = append(batch, entry)
			if len(batch) < batchSize {
				continue
			}
		case <-ticker.C:
			if len(batch) == 0 {
				continue
			}
		}

		if err := flush(); err != nil {
			return fmt.Errorf("failed while inserting batch: %w", err)
		}
	}
}
