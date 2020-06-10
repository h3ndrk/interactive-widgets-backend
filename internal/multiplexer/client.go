package multiplexer

import "github.com/h3ndrk/interactive-markdown/internal/id"

// Client represents a client connecting to a page instance.
type Client interface {
	// Read returns messages sent by the client. If there is no data available
	// at the moment, Read blocks until data is available again. If the client
	// disconnects, Read unblocks and returns io.EOF (also for future calls).
	Read() (id.WidgetID, []byte, error)

	// Write sends given data of a given widget to the client. If the client
	// cannot receive any data currently, Write blocks until the data has been
	// sent to the client. It is safe to call this function from different
	// goroutines.
	Write(id.WidgetID, []byte) error
}
