package docker

import (
	"bytes"
	"fmt"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type InstantiatedTextWidget struct {
	reader chan pages.Message
	writer chan pages.Message

	contents []byte
	errors   [][]byte
}

type TextMessage struct {
	Contents []byte   `json:"contents"`
	Errors   [][]byte `json:"errors"`
}

func NewInstantiatedTextWidget(widgetID pages.WidgetID, file string) (*InstantiatedTextWidget, error) {
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

	widget := &InstantiatedTextWidget{
		reader: make(chan pages.Message),
		writer: make(chan pages.Message),
	}
	go func() {
		for data := range process.OutputData {
			switch data.Origin {
			case StdoutStream:
				if bytes.Compare(data.Bytes, widget.contents) != 0 {
					widget.contents = data.Bytes
					widget.reader <- pages.Message{
						WidgetID: widgetID,
						Data: TextMessage{
							Contents: widget.contents,
							Errors:   widget.errors,
						},
					}
				}
			case StderrStream:
				widget.errors = append(widget.errors[:4], data.Bytes)
				widget.reader <- pages.Message{
					WidgetID: widgetID,
					Data: TextMessage{
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

func (d InstantiatedTextWidget) GetReader() <-chan pages.Message {
	return d.reader
}

func (d InstantiatedTextWidget) GetWriter() chan<- pages.Message {
	return d.writer
}
