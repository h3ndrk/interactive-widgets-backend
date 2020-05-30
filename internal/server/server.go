package server

import (
	"net/http"
)

// Server implements the http.Handler interface and represents a server
// interacting with a multiplexer.
type Server http.Handler
