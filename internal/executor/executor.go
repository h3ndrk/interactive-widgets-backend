package executor

import "github.com/h3ndrk/containerized-playground/internal/id"

// Executor instantiates pages and widgets in a execution backend. The methods
// regarding one page must not be used concurrently.
type Executor interface {
	// StartPage instantiates a page and the corresponding widgets. A page must
	// not be started more than once. It blocks while the instantiation is in
	// progress. Only if StartPage returns successfully, Read and Write
	// functions on the started widgets can be called.
	StartPage(id.PageID) error

	// StopPage tears down the page and its corresponding widgets. A page must
	// not be stopped more than once. A page must be running before calling
	// StopPage. It blocks while the teardown is in progress. All blocking Read
	// calls return during the execution of StopPage or shortly after. Read or
	// Write must not be called once StopPage has been called or is in progress.
	StopPage(id.PageID) error

	// Read returns data generated from the requested widget. If there is no
	// data available at the moment, Read blocks until data is available again.
	// If the widget stops (via StopPage), Read unblocks and returns io.EOF
	// (also for future calls). It is safe to call this function with same
	// arguments from different goroutines.
	Read(id.WidgetID) ([]byte, error)

	// Write sends data to the given widget. If the widget cannot receive any
	// data currently, Write blocks until the data has been sent to the widget.
	// It is safe to call this function with same widget ID arguments from
	// different goroutines.
	Write(id.WidgetID, []byte) error

	// GetCurrentState retrieves the current state from the given widget. It is
	// safe to call this function with same widget ID arguments from different
	// goroutines.
	GetCurrentState(id.WidgetID) ([]byte, error)
}
