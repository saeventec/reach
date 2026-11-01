package httpserver

import (
	"context"
	"net/http"
	"testing"
)

func TestRunStopsCleanlyOnTheStopSignal(t *testing.T) {
	// Port 0 asks the operating system for any free port.
	server := &http.Server{Addr: "127.0.0.1:0"}

	// A context that is already "done", as if the stop signal had arrived.
	stopSignal, stop := context.WithCancel(context.Background())
	stop()

	err := Run(stopSignal, "test", server)

	if err != nil {
		t.Errorf("got error %v, want none", err)
	}
}

func TestRunReportsAServerThatCannotStart(t *testing.T) {
	// No computer has a port -1, so the server fails at once.
	server := &http.Server{Addr: "127.0.0.1:-1"}

	// This stop signal never arrives, so only the server's error can end Run.
	stopSignal := context.Background()

	err := Run(stopSignal, "test", server)

	if err == nil {
		t.Error("got no error, want one")
	}
}
