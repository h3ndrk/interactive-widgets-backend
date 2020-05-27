package docker

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type InstantiatedButtonWidget struct {
	reader chan pages.OutgoingMessage
	writer chan pages.IncomingMessage

	mutex   sync.Mutex
	running bool
	stop    bool
}

type ButtonClickMessage struct {
	Click bool `json:"click"`
}

type ButtonOutputMessage struct {
	Origin string `json:"origin"`
	Data   []byte `json:"data"`
}

func NewInstantiatedButtonWidget(widgetID pages.WidgetID, command string) (*InstantiatedButtonWidget, error) {
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

	widget := &InstantiatedButtonWidget{
		reader: make(chan pages.OutgoingMessage),
		writer: make(chan pages.IncomingMessage),
	}
	go func() {
		for message := range widget.writer {
			var clickMessage ButtonClickMessage
			if err := json.Unmarshal(message.Data, &clickMessage); err == nil && clickMessage.Click {
				widget.mutex.Lock()
				if !widget.running {
					widget.running = true

					go func() {
						process, err := NewShortRunningProcess([]string{"docker", "run", "--rm", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), imageName, "/bin/bash", "-c", command})
						if err != nil {
							log.Print(errors.Wrapf(err, "Failed to execute button click command for widget %s", widgetID))
							return
						}

						for data := range process.OutputData {
							switch data.Origin {
							case StdoutStream:
								widget.reader <- pages.OutgoingMessage{
									WidgetID: widgetID,
									Data: ButtonOutputMessage{
										Origin: "stdout",
										Data:   data.Bytes,
									},
								}
							case StderrStream:
								widget.reader <- pages.OutgoingMessage{
									WidgetID: widgetID,
									Data: ButtonOutputMessage{
										Origin: "stderr",
										Data:   data.Bytes,
									},
								}
							}
						}

						process.Wait()

						widget.mutex.Lock()
						defer widget.mutex.Unlock()
						widget.running = false
						if widget.stop {
							close(widget.reader)
						}
					}()
				}

				widget.mutex.Unlock()

				continue
			}
		}

		// writer closed, gracefully terminate widget (close reader after process termination)
		widget.mutex.Lock()
		defer widget.mutex.Unlock()
		if widget.running {
			widget.stop = true
		} else {
			close(widget.reader)
		}
	}()

	return widget, nil
}

func (i InstantiatedButtonWidget) GetReader() <-chan pages.OutgoingMessage {
	return i.reader
}

func (i InstantiatedButtonWidget) GetWriter() chan<- pages.IncomingMessage {
	return i.writer
}
