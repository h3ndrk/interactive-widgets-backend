package docker

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/h3ndrk/containerized-playground/internal/executor"
	"github.com/h3ndrk/containerized-playground/internal/id"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// monitorWriteWidget represents one instance of a monitor-write widget (i.e.
// a widget that can monitor a file and can write to it) running as docker
// container. The process gets restarted if it stops but should not have
// stopped.
type monitorWriteWidget struct {
	stopWaiting   *sync.WaitGroup
	stopRequested bool
	connectWrite  bool

	runningMutex  sync.Mutex
	process       *exec.Cmd
	stdinWriter   io.WriteCloser
	stdoutChannel chan []byte
	stderrChannel chan []byte
	errChannel    chan error

	stateMutex sync.Mutex
	contents   []byte
	lastError  []byte
}

func newMonitorWriteWidget(widgetID id.WidgetID, file string, connectWrite bool) (widgetStream, error) {
	pageURL, roomID, _, err := id.PageURLAndRoomIDAndWidgetIndexFromWidgetID(widgetID)
	if err != nil {
		return nil, err
	}
	pageID, err := id.PageIDFromPageURLAndRoomID(pageURL, roomID)
	if err != nil {
		return nil, err
	}
	volumeName := fmt.Sprintf("containerized-playground-%s", id.EncodePageID(pageID))
	containerName := fmt.Sprintf("containerized-playground-%s", id.EncodeWidgetID(widgetID))

	w := &monitorWriteWidget{
		stopWaiting:  &sync.WaitGroup{},
		connectWrite: connectWrite,

		stdoutChannel: make(chan []byte),
		stderrChannel: make(chan []byte),
		errChannel:    make(chan error, 3), // ensure that errors are non-blocking (there are at most 3 places where at least one error gets written)
	}

	w.runningMutex.Lock()

	go func() {
		w.stopWaiting.Add(1)

		defer close(w.stdoutChannel)
		defer close(w.stderrChannel)
		defer close(w.errChannel) // errChannel needs to close last
		defer w.runningMutex.Unlock()
	loop:
		for {
			w.process = exec.Command("docker", "run", "--rm", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), "containerized-playground-monitor-write", file)

			stdinWriter, err := w.process.StdinPipe()
			if err != nil {
				log.Print(errors.Wrap(err, "Failed to create stdin pipe for monitor-write process"))
				if w.stopRequested {
					w.process = nil
					break loop
				} else {
					// restarting process
					time.Sleep(time.Second)
					continue
				}
			}
			w.stdinWriter = stdinWriter

			stdoutPipe, err := w.process.StdoutPipe()
			if err != nil {
				log.Print(errors.Wrap(err, "Failed to create stdout pipe for monitor-write process"))
				if w.stopRequested {
					w.process = nil
					break loop
				} else {
					// restarting process
					time.Sleep(time.Second)
					continue
				}
			}
			stdoutScanner := bufio.NewScanner(stdoutPipe)

			stderrPipe, err := w.process.StderrPipe()
			if err != nil {
				log.Print(errors.Wrap(err, "Failed to create stderr pipe for monitor-write process"))
				if w.stopRequested {
					w.process = nil
					break loop
				} else {
					// restarting process
					time.Sleep(time.Second)
					continue
				}
			}
			stderrScanner := bufio.NewScanner(stderrPipe)

			err = w.process.Start()
			if err != nil {
				log.Print(errors.Wrap(err, "Failed to start monitor-write process"))
				if w.stopRequested {
					w.process = nil
					break loop
				} else {
					// restarting process
					time.Sleep(time.Second)
					continue
				}
			}

			if !w.connectWrite {
				if err := stdinWriter.Close(); err != nil {
					log.Print(errors.Wrap(err, "Failed to close disconnected stdin pipe of monitor-write process"))
				}
			}

			var outputStreamWaiting sync.WaitGroup

			outputStreamWaiting.Add(1)
			go func() {
				defer outputStreamWaiting.Done()

				for stdoutScanner.Scan() {
					decoded, err := base64.StdEncoding.DecodeString(stdoutScanner.Text())
					if err != nil {
						w.errChannel <- err
						break
					}

					w.stdoutChannel <- decoded
				}
				if err := stdoutScanner.Err(); err != nil {
					w.errChannel <- err
				}
			}()

			outputStreamWaiting.Add(1)
			go func() {
				defer outputStreamWaiting.Done()

				for stderrScanner.Scan() {
					w.stderrChannel <- stderrScanner.Bytes()
				}
				if err := stderrScanner.Err(); err != nil {
					w.errChannel <- err
				}
			}()

			w.runningMutex.Unlock()

			w.process.Wait()
			outputStreamWaiting.Wait()

			w.runningMutex.Lock()

			if w.stopRequested {
				w.process = nil
				break loop
			} else {
				// restarting process
				time.Sleep(time.Second)
			}
		}
	}()

	return w, nil
}

// Read returns messages from the running monitor-write process.
func (w *monitorWriteWidget) Read() ([]byte, error) {
	stdoutCloseDetected := false
	stderrCloseDetected := false

	// try to read from stdout or stderr with higher priority (to drain stdout/stderr channels)
	for !stdoutCloseDetected || !stderrCloseDetected {
		if !stdoutCloseDetected && !stderrCloseDetected {
			select {
			case data, ok := <-w.stdoutChannel:
				if !ok {
					stdoutCloseDetected = true
				} else {
					w.stateMutex.Lock()

					if bytes.Compare(data, w.contents) != 0 {
						w.contents = data

						marshalled, err := json.Marshal(&executor.MonitorWriteContentsMessage{
							Contents: data,
						})
						if err != nil {
							log.Print(errors.Wrap(err, "Failed to marshal contents message"))
							time.Sleep(time.Second)
							// retry
						} else {
							w.stateMutex.Unlock()

							return marshalled, nil
						}
					}

					w.stateMutex.Unlock()
				}
			case data, ok := <-w.stderrChannel:
				if !ok {
					stderrCloseDetected = true
				} else {
					w.stateMutex.Lock()

					if bytes.Compare(data, w.lastError) != 0 {
						w.lastError = data

						marshalled, err := json.Marshal(&executor.MonitorWriteErrorMessage{
							Error: data,
						})
						if err != nil {
							log.Print(errors.Wrap(err, "Failed to marshal error message"))
							time.Sleep(time.Second)
							// retry
						} else {
							w.stateMutex.Unlock()

							return marshalled, nil
						}
					}

					w.stateMutex.Unlock()
				}
			}
		} else if !stdoutCloseDetected && stderrCloseDetected {
			select {
			case data, ok := <-w.stdoutChannel:
				if !ok {
					stdoutCloseDetected = true
				} else {
					w.stateMutex.Lock()

					if bytes.Compare(data, w.contents) != 0 {
						w.contents = data
						w.lastError = nil

						marshalled, err := json.Marshal(&executor.MonitorWriteContentsMessage{
							Contents: data,
						})
						if err != nil {
							log.Print(errors.Wrap(err, "Failed to marshal contents message"))
							time.Sleep(time.Second)
							// retry
						} else {
							w.stateMutex.Unlock()

							return marshalled, nil
						}
					}

					w.stateMutex.Unlock()
				}
			}
		} else if stdoutCloseDetected && !stderrCloseDetected {
			select {
			case data, ok := <-w.stderrChannel:
				if !ok {
					stderrCloseDetected = true
				} else {
					w.stateMutex.Lock()

					if bytes.Compare(data, w.lastError) != 0 {
						w.lastError = data

						marshalled, err := json.Marshal(&executor.MonitorWriteErrorMessage{
							Error: data,
						})
						if err != nil {
							log.Print(errors.Wrap(err, "Failed to marshal error message"))
							time.Sleep(time.Second)
							// retry
						} else {
							w.stateMutex.Unlock()

							return marshalled, nil
						}
					}

					w.stateMutex.Unlock()
				}
			}
		}
	}

	// at this point stdout and stderr are closed, therefore finish stopWaiting wait group
	defer w.stopWaiting.Done()

	// read all errors and return them as one cumulated error
	var cumulatedErrors error
	for err := range w.errChannel {
		cumulatedErrors = multierror.Append(cumulatedErrors, err)
	}
	if cumulatedErrors != nil {
		return nil, errors.Wrap(cumulatedErrors, "Failed to read from monitor-write process")
	}

	return nil, io.EOF
}

// Write writes messages to the running monitor-write process if stdin is
// connected.
func (w *monitorWriteWidget) Write(data []byte) error {
	if w.connectWrite {
		var inputMessage executor.MonitorWriteInputMessage
		if err := json.Unmarshal(data, &inputMessage); err != nil {
			return err
		}

		_, err := w.stdinWriter.Write(append(inputMessage.Contents, "\n"...))
		if err != nil {
			return errors.Wrap(err, "Failed to write data to monitor-write process")
		}
	}

	return nil
}

// Close stops the running monitor-write process. Afterwards, it waits for the
// process to terminate.
func (w *monitorWriteWidget) Close() error {
	w.runningMutex.Lock()
	process := w.process
	w.stopRequested = true
	w.runningMutex.Unlock()

	// this case is only skipped when Close is called after another Close and process termination
	if process != nil {
		if err := process.Process.Signal(syscall.SIGTERM); err != nil {
			return errors.Wrap(err, "Failed to send signal to monitor-write process")
		}
	}

	w.stopWaiting.Wait()

	return nil
}

// GetCurrentState returns an empty JSON object (there is no state).
func (w *monitorWriteWidget) GetCurrentState() ([]byte, error) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()

	if w.lastError == nil {
		marshalled, err := json.Marshal(&executor.MonitorWriteContentsMessage{
			Contents: w.contents,
		})
		if err != nil {
			return nil, errors.Wrap(err, "Failed to marshal contents message")
		}

		return marshalled, nil
	}

	marshalled, err := json.Marshal(&executor.MonitorWriteErrorMessage{
		Error: w.lastError,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal contents message")
	}

	return marshalled, nil
}
