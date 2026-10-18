package apis

import (
	"net/http"
	"net/http/pprof"

	"github.com/5733ccV10/go-log/internal/storage"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func NewRouter(queue chan<- storage.Entry, db driver.Conn) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /v1/logs", func(w http.ResponseWriter, r *http.Request) {
		logHandler(w, r, queue)
	})
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /v1/logs", func(w http.ResponseWriter, r *http.Request) {
		fetchLogsHandler(w, r, db)
	})
	return mux
}
