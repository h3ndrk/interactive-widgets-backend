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
	"github.com/pkg/errors"
)

// buttonWidget represents one instance of a button widget which can be invoked
// by a click. The click runs a defined command as docker container. Multiple
// clicks while a process is running are discarded (at most one process runs
// in parallel).
type buttonWidget struct {
	stopWaiting *sync.WaitGroup
	output      chan executor.ButtonOutputMessage
	clear       chan executor.ButtonClearMessage

	widgetID id.WidgetID
	command  string

	mutex         sync.Mutex
	stopRequested bool
	process       *exec.Cmd
}

func newButtonWidget(widgetID id.WidgetID, widget *parser.ButtonWidget) (widgetStream, error) {
	return &buttonWidget{
		stopWaiting: &sync.WaitGroup{},
		output:      make(chan executor.ButtonOutputMessage),
		clear:       make(chan executor.ButtonClearMessage),
		widgetID:    widgetID,
		command:     widget.Command,
	}, nil
}

// Read returns messages from the internal output or clear channel.
func (w *buttonWidget) Read() ([]byte, error) {
	select {
	case data, ok := <-w.output:
		if !ok {
			return nil, io.EOF
		}

		return json.Marshal(data)
	case data, ok := <-w.clear:
		if !ok {
			return nil, io.EOF
		}

		return json.Marshal(data)
	}
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

	if w.process == nil && !w.stopRequested {
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

		w.process = exec.Command("docker", "run", "--rm", "--name", containerName, "--network=none", "--memory=128m", "--cpus=0.1", "--pids-limit=128", "--cap-drop=ALL", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), imageName, "/bin/bash", "-c", w.command)

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

		w.clear <- executor.ButtonClearMessage{
			Clear: true,
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

			if err := w.process.Wait(); err != nil {
				w.output <- executor.ButtonOutputMessage{
					Origin: executor.StderrStream,
					Data:   []byte(errors.Wrapf(err, "Error while running button widget command").Error()),
				}
			}

			w.mutex.Lock()
			defer w.mutex.Unlock()
			w.process = nil
			if w.stopRequested {
				close(w.output)
				close(w.clear)
			}
		}()

	}

	return nil
}

// Close stops this widget by marking it as stop-requested and eventually
// sending SIGTERM to a running process. The output channel (from which Read
// reads) is closed either immediatly or after process termination. Afterwards,
// it waits for the process to terminate.
func (w *buttonWidget) Close() error {
	w.mutex.Lock()

	w.stopRequested = true

	if w.process != nil {
		if err := w.process.Process.Signal(syscall.SIGTERM); err != nil {
			return errors.Wrap(err, "Failed to send signal to button process")
		}
	} else {
		close(w.output)
		close(w.clear)
	}

	w.mutex.Unlock()

	w.stopWaiting.Wait()

	return nil
}

// GetCurrentState always returns an empty JSON object (there is no state).
func (w *buttonWidget) GetCurrentState() ([]byte, error) {
	return []byte("{}"), nil
}
