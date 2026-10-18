package apis

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/5733ccV10/go-log/internal/storage"
)

type ingestLogRequest struct {
	ServiceID  string            `json:"serviceID"`
	OccurredAt time.Time         `json:"occurredAt"`
	Message    string            `json:"message"`
	Level      storage.Level     `json:"level"`
	Metadata   map[string]string `json:"metadata"`
}

type postLogResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func logHandler(w http.ResponseWriter, r *http.Request, queue chan<- storage.Entry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	request := ingestLogRequest{}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		log.Printf("error while decoding: %v", err)
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Status: http.StatusText(http.StatusBadRequest),
			Error:  err.Error(),
		})
		return
	}

	if err := validateLogRequest(request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Status: http.StatusText(http.StatusBadRequest),
			Error:  err.Error(),
		})
		return
	}

	entry := storage.Entry{
		ServiceID:  request.ServiceID,
		OccurredAt: request.OccurredAt,
		ReceivedAt: time.Now().UTC(),
		Message:    request.Message,
		Level:      request.Level,
		Metadata:   request.Metadata,
	}

	if r.Context().Err() != nil {
		return
	}

	select {
	case queue <- entry:
	default:
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{
			Status: http.StatusText(http.StatusServiceUnavailable),
			Error:  "log queue is full; try again later",
		})
		return
	}

	writeJSON(w, http.StatusAccepted, postLogResponse{
		Status:  http.StatusText(http.StatusAccepted),
		Message: "log queued",
	})
}

func validateLogRequest(request ingestLogRequest) error {
	if strings.TrimSpace(request.ServiceID) == "" {
		return errors.New("serviceID is required")
	}
	if request.OccurredAt.IsZero() {
		return errors.New("occurredAt is required")
	}
	if strings.TrimSpace(request.Message) == "" {
		return errors.New("message is required")
	}
	switch request.Level {
	case storage.LevelDebug, storage.LevelInfo, storage.LevelWarn, storage.LevelError:
		return nil
	default:
		return errors.New("level must be DEBUG, INFO, WARN, or ERROR")
	}
}

