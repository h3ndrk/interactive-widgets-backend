package docker

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type InstantiatedPage struct {
	instantiatedWidgets []pages.InstantiatedWidget

	pageID pages.PageID

	reader chan pages.OutgoingMessage
	writer chan pages.IncomingMessage
}

func NewInstantiatedPage(pageID pages.PageID, widgets []pages.Widget) (*InstantiatedPage, error) {
	pageURL, roomID, err := pages.PageURLAndRoomIDFromPageID(pageID)
	if err != nil {
		return nil, err
	}
	volumeName := fmt.Sprintf("containerized-playground-%s", pages.EncodePageID(pageID))

	process, err := NewShortRunningProcess([]string{"docker", "volume", "create", volumeName})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create volume for page %s", pageID)
	}
	for data := range process.OutputData {
		switch data.Origin {
		case StdoutStream:
			fmt.Fprintln(os.Stdout, string(data.Bytes))
		case StderrStream:
			fmt.Fprintln(os.Stderr, string(data.Bytes))
		}
	}
	process.Wait()

	var instantiatedWidgets []pages.InstantiatedWidget
	for widgetIndex, widget := range widgets {
		widgetID, err := pages.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, pages.WidgetIndex(widgetIndex))
		if err != nil {
			// cleanup already created widgets
			for _, widgetToTeardown := range instantiatedWidgets {
				close(widgetToTeardown.GetWriter())
			}

			// remove volume
			process, err := NewShortRunningProcess([]string{"docker", "volume", "rm", volumeName})
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to remove volume for page %s", pageID)
			}
			for data := range process.OutputData {
				switch data.Origin {
				case StdoutStream:
					fmt.Fprintln(os.Stdout, string(data.Bytes))
				case StderrStream:
					fmt.Fprintln(os.Stderr, string(data.Bytes))
				}
			}
			process.Wait()

			// return error
			return nil, errors.Wrapf(err, "Failed to instantiated widget %d", widgetIndex)
		}

		widget, err := widget.Instantiate(widgetID)
		if err != nil {
			// cleanup already created widgets
			for _, widgetToTeardown := range instantiatedWidgets {
				close(widgetToTeardown.GetWriter())
			}

			// remove volume
			process, err := NewShortRunningProcess([]string{"docker", "volume", "rm", volumeName})
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to remove volume for page %s", pageID)
			}
			for data := range process.OutputData {
				switch data.Origin {
				case StdoutStream:
					fmt.Fprintln(os.Stdout, string(data.Bytes))
				case StderrStream:
					fmt.Fprintln(os.Stderr, string(data.Bytes))
				}
			}
			process.Wait()

			// return error
			return nil, errors.Wrapf(err, "Failed to instantiated widget %d", widgetIndex)
		}

		instantiatedWidgets = append(instantiatedWidgets, widget)
	}

	reader := make(chan pages.OutgoingMessage)
	writer := make(chan pages.IncomingMessage)
	var closeWaiting sync.WaitGroup

	closeWaiting.Add(len(instantiatedWidgets))
	for _, widget := range instantiatedWidgets {
		go func(widget pages.InstantiatedWidget) {
			// receive all messages from widgets
			for message := range widget.GetReader() {
				reader <- message
			}

			// inform wait group that another widget closed it's reader
			closeWaiting.Done()
		}(widget)
	}

	go func(closeWaiting *sync.WaitGroup) {
		// wait until all widgets closed their readers
		closeWaiting.Wait()

		// remove volume
		process, err := NewShortRunningProcess([]string{"docker", "volume", "rm", volumeName})
		if err != nil {
			log.Print(errors.Wrapf(err, "Failed to remove volume for page %s", pageID))
		} else {
			for data := range process.OutputData {
				switch data.Origin {
				case StdoutStream:
					fmt.Fprintln(os.Stdout, string(data.Bytes))
				case StderrStream:
					fmt.Fprintln(os.Stderr, string(data.Bytes))
				}
			}
			process.Wait()
		}

		// close page's reader
		close(reader)
	}(&closeWaiting)

	go func() {
		for message := range writer {
			_, _, widgetIndex, err := pages.PageURLAndRoomIDAndWidgetIndexFromWidgetID(message.WidgetID)
			if err != nil {
				log.Print(err)
				continue
			}
			if widgetIndex < 0 || int(widgetIndex) >= len(instantiatedWidgets) {
				log.Printf("Failed to forward message to widget ID %s: widget index out of bounds", message.WidgetID)
				continue
			}

			instantiatedWidgets[widgetIndex].GetWriter() <- message
		}

		// writer closed, close widget writers
		for _, widget := range instantiatedWidgets {
			close(widget.GetWriter())
		}
	}()

	return &InstantiatedPage{
		instantiatedWidgets: instantiatedWidgets,
		pageID:              pageID,
		reader:              reader,
		writer:              writer,
	}, nil
}

func (d InstantiatedPage) GetReader() <-chan pages.OutgoingMessage {
	return d.reader
}

func (d InstantiatedPage) GetWriter() chan<- pages.IncomingMessage {
	return d.writer
}
