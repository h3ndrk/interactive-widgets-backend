package docker

import (
	"bytes"
	"fmt"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type InstantiatedImageWidget struct {
	reader chan pages.OutgoingMessage
	writer chan pages.IncomingMessage

	contents []byte
	errors   [][]byte
}

type ImageContentsMessage struct {
	Contents []byte   `json:"contents"`
	Errors   [][]byte `json:"errors"`
}

func NewInstantiatedImageWidget(widgetID pages.WidgetID, file string) (*InstantiatedImageWidget, error) {
	pageURL, roomID, _, err := pages.PageURLAndRoomIDAndWidgetIndexFromWidgetID(widgetID)
	if err != nil {
		return nil, err
	}
	pageID, err := pages.PageIDFromPageURLAndRoomID(pageURL, roomID)
	if err != nil {
		return nil, err
	}
	volumeName := fmt.Sprintf("containerized-playground-%s", pages.EncodePageID(pageID))
	containerName := fmt.Sprintf("containerized-playground-%s", pages.EncodeWidgetID(widgetID))

	process, err := NewLongRunningProcess([]string{"docker", "run", "--rm", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), "containerized-playground-monitor-write", file}, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to run container for widget %s", widgetID)
	}

	widget := &InstantiatedImageWidget{
		reader: make(chan pages.OutgoingMessage),
		writer: make(chan pages.IncomingMessage),
	}
	go func() {
		for data := range process.OutputData {
			switch data.Origin {
			case StdoutStream:
				if bytes.Compare(data.Bytes, widget.contents) != 0 {
					widget.contents = data.Bytes
					widget.reader <- pages.OutgoingMessage{
						WidgetID: widgetID,
						Data: ImageContentsMessage{
							Contents: widget.contents,
							Errors:   widget.errors,
						},
					}
				}
			case StderrStream:
				widget.errors = append(widget.errors[:4], data.Bytes)
				widget.reader <- pages.OutgoingMessage{
					WidgetID: widgetID,
					Data: ImageContentsMessage{
						Contents: widget.contents,
						Errors:   widget.errors,
					},
				}
			}
		}

		// process stopped, close reader
		close(widget.reader)
	}()
	go func() {
		for range widget.writer {
			// discard
		}

		// writer closed, stop process
		process.Stop()
	}()

	return widget, nil
}

func (i InstantiatedImageWidget) GetReader() <-chan pages.OutgoingMessage {
	return i.reader
}

func (i InstantiatedImageWidget) GetWriter() chan<- pages.IncomingMessage {
	return i.writer
}
