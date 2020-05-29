package docker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"

	"github.com/h3ndrk/containerized-playground/internal/executor"
	"github.com/h3ndrk/containerized-playground/internal/id"
	"github.com/h3ndrk/containerized-playground/internal/parser"
)

type buttonWidget struct {
	stopWaiting *sync.WaitGroup
	output      chan executor.ButtonOutputMessage

	widgetID id.WidgetID
	command  string

	mutex   sync.Mutex
	running bool
	stop    bool
	process *exec.Cmd
}

func newButtonWidget(widgetID id.WidgetID, widget parser.ButtonWidget) (widgetStream, error) {
	return &buttonWidget{
		output:   make(chan executor.ButtonOutputMessage),
		widgetID: widgetID,
		command:  widget.Command,
	}, nil
}

func (w *buttonWidget) Read() ([]byte, error) {
	data, ok := <-w.output
	if !ok {
		return nil, io.EOF
	}

	return json.Marshal(data)
}

func (w *buttonWidget) Write(data []byte) error {
	var inputMessage executor.ButtonClickMessage
	if err := json.Unmarshal(data, &inputMessage); err != nil {
		return err
	}
	if !inputMessage.Click {
		// discard
		return nil
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.running {
		w.running = true

		pageURL, roomID, _, err := id.PageURLAndRoomIDAndWidgetIndexFromWidgetID(w.widgetID)
		if err != nil {
			return err
		}
		pageID, err := id.PageIDFromPageURLAndRoomID(pageURL, roomID)
		if err != nil {
			return err
		}
		volumeName := fmt.Sprintf("containerized-playground-%s", id.EncodePageID(pageID))
		imageName := fmt.Sprintf("containerized-playground-%s", id.EncodePageURL(pageURL))
		containerName := fmt.Sprintf("containerized-playground-%s", id.EncodeWidgetID(widgetID))

		go func() {
			w.stopWaiting.Add(1)
			defer w.stopWaiting.Done()

			w.process = exec.Command("docker", "run", "--rm", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), imageName, "/bin/bash", "-c", w.command)

			stdoutPipe, err := w.process.StdoutPipe()
			if err != nil {
				w.output <- executor.ButtonOutputMessage{
					Origin: executor.StderrStream,
					Data:   []byte(err.Error()),
				}
				return
			}
			stdoutScanner := bufio.NewScanner(stdoutPipe)

			stderrPipe, err := w.process.StderrPipe()
			if err != nil {
				w.output <- executor.ButtonOutputMessage{
					Origin: executor.StderrStream,
					Data:   []byte(err.Error()),
				}
				return
			}
			stderrScanner := bufio.NewScanner(stderrPipe)

			err = w.process.Start()
			if err != nil {
				w.output <- executor.ButtonOutputMessage{
					Origin: executor.StderrStream,
					Data:   []byte(err.Error()),
				}
				return
			}

			go func() {
				for stdoutScanner.Scan() {
					w.output <- executor.ButtonOutputMessage{
						Origin: executor.StdoutStream,
						Data:   stdoutScanner.Bytes(),
					}
				}
			}()

			go func() {
				for stderrScanner.Scan() {
					w.output <- executor.ButtonOutputMessage{
						Origin: executor.StderrStream,
						Data:   stderrScanner.Bytes(),
					}
				}
			}()

			w.process.Wait()

			w.mutex.Lock()
			defer w.mutex.Unlock()
			w.running = false
			w.process = nil
			if w.stop {
				close(w.output)
			}
		}()

	}

	return nil
}

func (w *buttonWidget) Close() {
	w.mutex.Lock()

	w.stop = true

	if w.process != nil {
		if err := w.process.Process.Signal(syscall.SIGTERM); err != nil {
			w.output <- executor.ButtonOutputMessage{
				Origin: executor.StderrStream,
				Data:   []byte(err.Error()),
			}
		}
	} else {
		close(w.output)
	}

	w.mutex.Unlock()

	w.stopWaiting.Wait()

	return nil
}
