package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/h3ndrk/containerized-playground/internal/executor"
	"github.com/h3ndrk/containerized-playground/internal/id"
	"github.com/h3ndrk/containerized-playground/internal/parser"
	"github.com/pkg/errors"
)

// terminalWidget represents one instance of a terminal widget running as
// docker container. A terminal widget runs a process in a pseudo terminal. The
// process gets restarted if it stops but should not have stopped.
type terminalWidget struct {
	stopWaiting   *sync.WaitGroup
	stopRequested bool

	runningMutex    sync.Mutex
	process         *exec.Cmd
	pseudoTerminal  *os.File
	sharedReadChunk []byte

	mutex    sync.Mutex
	contents []byte
	errors   [][]byte
}

func newTerminalWidget(widgetID id.WidgetID, widget *parser.TerminalWidget) (widgetStream, error) {
	pageURL, roomID, _, err := id.PageURLAndRoomIDAndWidgetIndexFromWidgetID(widgetID)
	if err != nil {
		return nil, err
	}
	pageID, err := id.PageIDFromPageURLAndRoomID(pageURL, roomID)
	if err != nil {
		return nil, err
	}
	volumeName := fmt.Sprintf("containerized-playground-%s", id.EncodePageID(pageID))
	imageName := fmt.Sprintf("containerized-playground-%s", id.EncodePageURL(pageURL))
	containerName := fmt.Sprintf("containerized-playground-%s", id.EncodeWidgetID(widgetID))

	w := &terminalWidget{
		stopWaiting:     &sync.WaitGroup{},
		sharedReadChunk: make([]byte, 4096),
	}

	w.runningMutex.Lock()

	go func() {
		w.stopWaiting.Add(1)
		defer w.stopWaiting.Done()
	loop:
		for {
			w.process = exec.Command("docker", "run", "--rm", "--interactive", "--tty", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), "--workdir", widget.WorkingDirectory, imageName, "/bin/bash")
			pseudoTerminal, err := pty.Start(w.process)
			if err != nil {
				if w.stopRequested {
					w.process = nil
					w.pseudoTerminal = nil
					// unlock to allow future reads and writes
					w.runningMutex.Unlock()
					break loop
				} else {
					// restarting process
					time.Sleep(time.Second)
					continue
				}
			}
			w.pseudoTerminal = pseudoTerminal
			w.runningMutex.Unlock()

			w.process.Wait()
			w.pseudoTerminal.Close()

			w.runningMutex.Lock()

			if w.stopRequested {
				w.process = nil
				w.pseudoTerminal = nil
				// unlock to allow future reads and writes
				w.runningMutex.Unlock()
				break loop
			} else {
				// restarting process
				time.Sleep(time.Second)
			}
		}
	}()

	return w, nil
}

// Read returns messages from the running pseudo terminal process.
func (w *terminalWidget) Read() ([]byte, error) {
	sharedReadChunkLength := 0
	for {
		w.runningMutex.Lock()
		pseudoTerminal := w.pseudoTerminal
		w.runningMutex.Unlock()

		if pseudoTerminal == nil {
			// this case only happens when Read is called after Close and process termination
			return nil, io.EOF
		}

		n, err := pseudoTerminal.Read(w.sharedReadChunk)
		if err != nil {
			if err == io.EOF || errors.Is(err, os.ErrClosed) {
				continue
			}

			return nil, errors.Wrap(err, "Failed to read from pseudo terminal process")
		}
		sharedReadChunkLength = n

		break
	}

	data, err := json.Marshal(&executor.TerminalOutputMessage{
		Data: w.sharedReadChunk[:sharedReadChunkLength],
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal output message")
	}

	return data, nil
}

// Write writes messages to the running pseudo terminal process.
func (w *terminalWidget) Write(data []byte) error {
	var inputMessage executor.TerminalInputMessage
	if err := json.Unmarshal(data, &inputMessage); err != nil {
		return errors.Wrap(err, "Failed to unmarshal input message")
	}

	w.runningMutex.Lock()
	pseudoTerminal := w.pseudoTerminal
	w.runningMutex.Unlock()

	// this case is only skipped when Write is called after Close and process termination
	if pseudoTerminal != nil {
		_, err := pseudoTerminal.Write(inputMessage.Data)
		if err != nil {
			return errors.Wrap(err, "Failed to write data to pseudo terminal")
		}
	}

	return nil
}

// Close closes the running pseudo terminal process. Afterwards, it waits for the
// process to terminate.
func (w *terminalWidget) Close() error {
	w.runningMutex.Lock()
	process := w.process
	w.stopRequested = true
	w.runningMutex.Unlock()

	// this case is only skipped when Close is called after another Close and process termination
	if process != nil {
		if err := process.Process.Signal(syscall.SIGTERM); err != nil {
			return errors.Wrap(err, "Failed to send signal to pseudo terminal process")
		}
	}

	w.stopWaiting.Wait()

	return nil
}
