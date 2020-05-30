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

// buttonWidget represents one instance of a button widget which can be invoked
// by a click. The click runs a defined command as docker container. Multiple
// clicks while a process is running are discarded (at most one process runs
// in parallel). Implementation errors are not passed to the output. The
// implementation communicates via one channel.
type buttonWidget struct {
	stopWaiting *sync.WaitGroup
	output      chan executor.ButtonOutputMessage

	widgetID id.WidgetID
	command  string

	mutex         sync.Mutex
	stopRequested bool
	process       *exec.Cmd
}

func newButtonWidget(widgetID id.WidgetID, widget parser.ButtonWidget) (widgetStream, error) {
	return &buttonWidget{
		output:   make(chan executor.ButtonOutputMessage),
		widgetID: widgetID,
		command:  widget.Command,
	}, nil
}

// Read returns messages from the internal output channel.
func (w *buttonWidget) Read() ([]byte, error) {
	data, ok := <-w.output
	if !ok {
		return nil, io.EOF
	}

	return json.Marshal(data)
}

// Write parses the given message and initiates a button click by starting the defined process.
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

	if w.process == nil {
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
		containerName := fmt.Sprintf("containerized-playground-%s", id.EncodeWidgetID(w.widgetID))

		w.stopWaiting.Add(1)

		w.process = exec.Command("docker", "run", "--rm", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), imageName, "/bin/bash", "-c", w.command)

		stdoutPipe, err := w.process.StdoutPipe()
		if err != nil {
			return err
		}
		stdoutScanner := bufio.NewScanner(stdoutPipe)

		stderrPipe, err := w.process.StderrPipe()
		if err != nil {
			return err
		}
		stderrScanner := bufio.NewScanner(stderrPipe)

		err = w.process.Start()
		if err != nil {
			return err
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

		go func() {
			defer w.stopWaiting.Done()

			w.process.Wait()

			w.mutex.Lock()
			defer w.mutex.Unlock()
			w.process = nil
			if w.stopRequested {
				close(w.output)
			}
		}()

	}

	return nil
}

// Close stops this widget by marking it as stop-requested and eventually
// sending SIGTERM to a running process. The output channel (from which Read
// reads) is closed either immediatly or after process termination.
func (w *buttonWidget) Close() {
	w.mutex.Lock()

	w.stopRequested = true

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
}
