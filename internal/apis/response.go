package apis

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func writeJSON(w http.ResponseWriter, statusCode int, response any) {
	body, err := json.Marshal(response)
	if err != nil {
		log.Printf("error while encoding: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		log.Printf("error while writing response: %v", err)
	}
}
