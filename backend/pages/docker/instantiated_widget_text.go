package docker

import (
	"fmt"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type DockerInstantiatedWidgetText struct {
	Reader chan pages.Message
	Writer chan pages.Message
}

type TextMessage struct {
	Origin string `json:"origin"`
	Bytes  []byte `json:"bytes"`
}

func NewDockerInstantiatedWidgetText(widgetID pages.WidgetID, file string) (*DockerInstantiatedWidgetText, error) {
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

	reader := make(chan pages.Message)
	writer := make(chan pages.Message)
	go func() {
		for data := range process.OutputData {
			var origin string
			switch data.Origin {
			case StdoutStream:
				origin = "stdout"
			case StderrStream:
				origin = "stderr"
			}
			reader <- pages.Message{
				WidgetID: widgetID,
				Data: TextMessage{
					Origin: origin,
					Bytes:  data.Bytes,
				},
			}
		}
		// process stopped, close reader
		close(reader)
	}()
	go func() {
		for range writer {
			// discard
		}
		// writer closed, stop process
		process.Stop()
	}()

	return &DockerInstantiatedWidgetText{
		Reader: reader,
		Writer: writer,
	}, nil
}

func (d DockerInstantiatedWidgetText) GetReader() <-chan pages.Message {
	return d.Reader
}

func (d DockerInstantiatedWidgetText) GetWriter() chan<- pages.Message {
	return d.Writer
}
