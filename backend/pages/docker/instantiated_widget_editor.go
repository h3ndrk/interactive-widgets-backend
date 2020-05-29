package docker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type InstantiatedEditorWidget struct {
	reader chan pages.OutgoingMessage
	writer chan pages.IncomingMessage

	contents []byte
	errors   [][]byte
}

type EditorInputMessage struct {
	Contents []byte `json:"contents"`
}

type EditorContentsMessage struct {
	Contents []byte   `json:"contents"`
	Errors   [][]byte `json:"errors"`
}

func NewInstantiatedEditorWidget(widgetID pages.WidgetID, file string) (*InstantiatedEditorWidget, error) {
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

	stdin := make(chan []byte)
	process, err := NewLongRunningProcess([]string{"docker", "run", "--rm", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), "containerized-playground-monitor-write", file}, stdin)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to run container for widget %s", widgetID)
	}

	widget := &InstantiatedEditorWidget{
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
						Data: EditorContentsMessage{
							Contents: widget.contents,
							Errors:   widget.errors,
						},
					}
				}
			case StderrStream:
				if len(widget.errors) > 4 {
					widget.errors = append(widget.errors[len(widget.errors)-4:len(widget.errors)], data.Bytes)
				} else {
					widget.errors = append(widget.errors, data.Bytes)
				}
				widget.reader <- pages.OutgoingMessage{
					WidgetID: widgetID,
					Data: EditorContentsMessage{
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
		for message := range widget.writer {
			var inputMessage EditorInputMessage
			if err := json.Unmarshal(message.Data, &inputMessage); err == nil {
				stdin <- []byte(base64.StdEncoding.EncodeToString(inputMessage.Contents))

				continue
			}
		}

		// writer closed, close stdin and stop process
		close(stdin)
		process.Stop()
	}()

	return widget, nil
}

func (i InstantiatedEditorWidget) GetReader() <-chan pages.OutgoingMessage {
	return i.reader
}

func (i InstantiatedEditorWidget) GetWriter() chan<- pages.IncomingMessage {
	return i.writer
}
