package shim

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"tinfoil/internal/boot"
)

func containersHandler() http.HandlerFunc {
	return serveContainerStatusFile(boot.ContainerStatusPath)
}

func serveContainerStatusFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "container status not available", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}
}

// NoContainers answers for an image that runs its workload without containers.
func NoContainers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			ObservedAt time.Time  `json:"observed_at"`
			Containers []struct{} `json:"containers"`
		}{time.Now().UTC(), []struct{}{}})
	}
}
