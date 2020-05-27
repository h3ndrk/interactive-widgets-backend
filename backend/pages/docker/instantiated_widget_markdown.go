package docker

import (
	"github.com/h3ndrk/containerized-playground/backend/pages"
)

type InstantiatedMarkdownWidget struct {
	reader chan pages.Message
	writer chan pages.Message
}

func NewInstantiatedMarkdownWidget(widgetID pages.WidgetID, file string) (*InstantiatedMarkdownWidget, error) {
	reader := make(chan pages.Message)
	writer := make(chan pages.Message)
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

func (i InstantiatedMarkdownWidget) GetReader() <-chan pages.Message {
	return i.reader
}

func (i InstantiatedMarkdownWidget) GetWriter() chan<- pages.Message {
	return i.writer
}
