package server

import (
	"net/http"
)

// Server implements the http.Handler interface and represents a server
// interacting with a multiplexer.
type Server interface {
	http.Handler

	// Shutdown terminates all open connections. It does not wait for
	// termination.
	Shutdown()

	// Wait waits for all connections to terminate.
	Wait()
}
