package docker

import (
	"encoding/json"
	"fmt"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type InstantiatedTerminalWidget struct {
	reader chan pages.OutgoingMessage
	writer chan pages.IncomingMessage

	contents []byte
	errors   [][]byte
}

type TerminalInputMessage struct {
	Data []byte `json:"data"`
}

type TerminalOutputMessage struct {
	Data []byte `json:"data"`
}

func NewInstantiatedTerminalWidget(widgetID pages.WidgetID, workingDirectory string) (*InstantiatedTerminalWidget, error) {
	pageURL, roomID, _, err := pages.PageURLAndRoomIDAndWidgetIndexFromWidgetID(widgetID)
	if err != nil {
		return nil, err
	}
	pageID, err := pages.PageIDFromPageURLAndRoomID(pageURL, roomID)
	if err != nil {
		return nil, err
	}
	volumeName := fmt.Sprintf("containerized-playground-%s", pages.EncodePageID(pageID))
	imageName := fmt.Sprintf("containerized-playground-%s", pages.EncodePageURL(pageURL))
	containerName := fmt.Sprintf("containerized-playground-%s", pages.EncodeWidgetID(widgetID))

	stdin := make(chan []byte)
	process, err := NewLongRunningTerminalProcess([]string{"docker", "run", "--rm", "--interactive", "--tty", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), "--workdir", workingDirectory, imageName, "/bin/bash"}, stdin)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to run container for widget %s", widgetID)
	}

	widget := &InstantiatedTerminalWidget{
		reader: make(chan pages.OutgoingMessage),
		writer: make(chan pages.IncomingMessage),
	}
	go func() {
		for data := range process.Output {
			widget.reader <- pages.OutgoingMessage{
				WidgetID: widgetID,
				Data: TerminalOutputMessage{
					Data: data,
				},
			}
		}

		// process stopped, close reader
		close(widget.reader)
	}()
	go func() {
		for message := range widget.writer {
			var inputMessage TerminalInputMessage
			if err := json.Unmarshal(message.Data, &inputMessage); err == nil {
				stdin <- inputMessage.Data

				continue
			}
		}

		// writer closed, close stdin and stop process
		close(stdin)
		process.Stop()
	}()

	return widget, nil
}

func (i InstantiatedTerminalWidget) GetReader() <-chan pages.OutgoingMessage {
	return i.reader
}

func (i InstantiatedTerminalWidget) GetWriter() chan<- pages.IncomingMessage {
	return i.writer
}
