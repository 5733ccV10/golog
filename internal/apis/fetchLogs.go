package apis

import (
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/5733ccV10/go-log/internal/storage"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type getLog struct {
	ServiceID  string            `json:"serviceID"`
	OccurredAt time.Time         `json:"occurredAt"`
	Message    string            `json:"message"`
	Level      storage.Level     `json:"level"`
	MetaData   map[string]string `json:"metadata"`
}

type getLogsResponse struct {
	Logs []getLog `json:"logs"`
}

func fetchLogsHandler(w http.ResponseWriter, r *http.Request, db driver.Conn) {
	query := r.URL.Query()

	from, err := time.Parse(time.RFC3339, query.Get("from"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Status: http.StatusText(http.StatusBadRequest),
			Error:  "from is required and must be an RFC3339 timestamp",
		})
		return
	}

	till, err := time.Parse(time.RFC3339, query.Get("till"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Status: http.StatusText(http.StatusBadRequest),
			Error:  "till is required and must be an RFC3339 timestamp",
		})
		return
	}
	count := 100
	countText := query.Get("count")

	if countText != "" {
		parsed, err := strconv.Atoi(countText)
		if err != nil || parsed < 1 || parsed > 1000 {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Status: http.StatusText(http.StatusBadRequest),
				Error:  "count must be an integer between 1 and 1000",
			})
			return
		}
		count = parsed
	}
	filter := storage.LogFilter{
		ServiceID: query.Get("serviceID"),
		Level:     storage.Level(query.Get("level")),
		From:      from,
		Till:      till,
		Count:     count,
	}

	if filter.From.After(filter.Till) {
		log.Printf("from is after till")
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Status: http.StatusText(http.StatusBadRequest),
			Error:  "from value is greater than till value",
		})
		return
	}

	allowedlevels := []storage.Level{
		storage.LevelDebug,
		storage.LevelInfo,
		storage.LevelWarn,
		storage.LevelError,
	}

	if filter.Level != "" && !slices.Contains(allowedlevels, filter.Level) {
		log.Printf("request log level not found")
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Status: http.StatusText(http.StatusBadRequest),
			Error:  "requested log levels does not exist",
		})
		return
	}

	response, err := storage.GetFilteredLogs(r.Context(), db, filter)

	if err != nil {
		log.Printf("something went wrong: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Status: http.StatusText(http.StatusInternalServerError),
			Error:  "could not fetch logs",
		})
		return
	}

	logs := make([]getLog, 0, filter.Count)

	for _, entry := range response {
		log := getLog{
			ServiceID:  entry.ServiceID,
			OccurredAt: entry.OccurredAt,
			Message:    entry.Message,
			Level:      entry.Level,
			MetaData:   entry.Metadata,
		}

		logs = append(logs, log)
	}

	logsResponse := getLogsResponse{
		Logs: logs,
	}
	writeJSON(w, http.StatusOK, logsResponse)
}
