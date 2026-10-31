package main

import (
	"io"
	"log/slog"
	"net/http"
)

// handleHealth answers "ok" so other tools can check that the service is
// running. Docker Compose (and later Kubernetes) calls it regularly.
// Learn more: docs/learn/http.md#health-checks
func handleHealth(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	_, err := io.WriteString(writer, "ok")
	if err != nil {
		// The caller went away before reading the answer. We can't fix
		// that, but we log it instead of hiding it.
		slog.Warn("could not write the health response", "error", err)
	}
}
