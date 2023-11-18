package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"sreda/internal/models"
)

func ProcessRequest(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entry models.IterationEntry
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&entry); err != nil {
			log.Error("Request error", "error", err.Error(), "status", http.StatusBadRequest)
			http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if entry.Iteration == 0 {
			log.Error("Request error", "error", "iteration is zero", "status", http.StatusTeapot)
			http.Error(w, "Failed to decode request body: iteration is zero", http.StatusTeapot)
			return
		}

		log.Debug("Request received", "iteration", entry.Iteration)

		w.WriteHeader(http.StatusOK)
	}
}
