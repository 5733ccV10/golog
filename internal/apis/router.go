package apis

import (
	"net/http"
	"net/http/pprof"

	"github.com/5733ccV10/go-log/internal/storage"
)

func NewRouter(queue chan<- storage.Entry) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /v1/logs", func(w http.ResponseWriter, r *http.Request) {
		logHandler(w, r, queue)
	})
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	return mux
}
