package docker

import (
	"github.com/h3ndrk/containerized-playground/backend/pages"
)

type InstantiatedMarkdownWidget struct {
	reader chan pages.OutgoingMessage
	writer chan pages.IncomingMessage
}

func NewInstantiatedMarkdownWidget(widgetID pages.WidgetID, file string) (*InstantiatedMarkdownWidget, error) {
	reader := make(chan pages.OutgoingMessage)
	writer := make(chan pages.IncomingMessage)
	go func() {
		for range writer {
			// discard
		}

		// writer closed, close reader
		close(reader)
	}()

	return &InstantiatedMarkdownWidget{
		reader: reader,
		writer: writer,
	}, nil
}

func (i InstantiatedMarkdownWidget) GetReader() <-chan pages.OutgoingMessage {
	return i.reader
}

func (i InstantiatedMarkdownWidget) GetWriter() chan<- pages.IncomingMessage {
	return i.writer
}
