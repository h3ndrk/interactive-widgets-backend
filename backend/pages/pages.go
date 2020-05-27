package pages

type IncomingMessage struct {
	WidgetID WidgetID
	Data     []byte
}

type OutgoingMessage struct {
	WidgetID WidgetID
	Data     interface{}
}

type ReadWriter struct {
	Reader <-chan IncomingMessage
	Writer chan<- OutgoingMessage
}

type Pages interface {
	Prepare() error
	Cleanup() error

	// Observe stores the observer and ensures that an instantiated page exists.
	// When adding an observer: Increase number of observers, if page did not exist before, instantiate it
	// When an observer closes: Decrease number of observers, if observers count zero, close page
	Observe(pageID PageID, observer ReadWriter) error

	MarshalPages() ([]byte, error)
	MarshalPage(pageURL PageURL) ([]byte, error)
}

type Page interface {
	Prepare() error
	Cleanup() error

	Instantiate(pageID PageID) (InstantiatedPage, error)

	MarshalPage() ([]byte, error)
	MarshalWidgets() ([]byte, error)
}

type Widget interface {
	Prepare() error
	Cleanup() error

	Instantiate(widgetID WidgetID) (InstantiatedWidget, error)

	MarshalWidget() ([]byte, error)
}

type InstantiatedPage interface {
	GetReader() <-chan OutgoingMessage
	GetWriter() chan<- IncomingMessage
}

type InstantiatedWidget interface {
	GetReader() <-chan OutgoingMessage
	GetWriter() chan<- IncomingMessage
}
