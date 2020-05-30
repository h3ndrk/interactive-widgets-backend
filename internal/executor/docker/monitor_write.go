package docker

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/h3ndrk/containerized-playground/internal/executor"
	"github.com/h3ndrk/containerized-playground/internal/id"
)

type monitorWriteWidget struct {
	running      chan struct{}
	stopWaiting  *sync.WaitGroup
	input        chan executor.MonitorWriteInputMessage
	output       chan executor.MonitorWriteOutputMessage
	connectWrite bool

	mutex    sync.Mutex
	contents []byte
	errors   [][]byte
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
		running:      make(chan struct{}),
		stopWaiting:  &sync.WaitGroup{},
		output:       make(chan executor.MonitorWriteOutputMessage),
		connectWrite: connectWrite,
	}

	if w.connectWrite {
		w.input = make(chan executor.MonitorWriteInputMessage)
	}

	go func() {
		w.stopWaiting.Add(1)
	loop:
		for {
			done := make(chan struct{})
			process := exec.Command("docker", "run", "--rm", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), "containerized-playground-monitor-write", file)

			stdinWriter, err := process.StdinPipe()
			if err != nil {
				w.storeAndSendError([]byte(err.Error()))
				time.Sleep(time.Second)
				continue
			}

			stdoutPipe, err := process.StdoutPipe()
			if err != nil {
				w.storeAndSendError([]byte(err.Error()))
				time.Sleep(time.Second)
				continue
			}
			stdoutScanner := bufio.NewScanner(stdoutPipe)

			stderrPipe, err := process.StderrPipe()
			if err != nil {
				w.storeAndSendError([]byte(err.Error()))
				time.Sleep(time.Second)
				continue
			}
			stderrScanner := bufio.NewScanner(stderrPipe)

			err = process.Start()
			if err != nil {
				w.storeAndSendError([]byte(err.Error()))
				time.Sleep(time.Second)
				continue
			}

			if w.connectWrite {
				go func() {
					defer stdinWriter.Close()

					for {
						select {
						case message, ok := <-w.input:
							if !ok {
								return
							}

							_, err := stdinWriter.Write(append(message.Contents, "\n"...))
							if err != nil {
								continue
							}
						case <-done:
							return
						}
					}
				}()
			} else {
				stdinWriter.Close()
			}

			go func() {
				for stdoutScanner.Scan() {
					decoded, err := base64.StdEncoding.DecodeString(stdoutScanner.Text())
					if err != nil {
						w.storeAndSendError([]byte(err.Error()))
						continue
					}

					w.storeAndSendContents(decoded)
				}
			}()

			go func() {
				for stderrScanner.Scan() {
					w.storeAndSendError(stderrScanner.Bytes())
				}
			}()

			go func() {
				select {
				case <-done:
				case _, ok := <-w.running:
					if !ok {
						if err := process.Process.Signal(syscall.SIGTERM); err != nil {
							w.storeAndSendError([]byte(err.Error()))
						}
					}
				}
			}()

			process.Wait()
			close(done)

			select {
			case _, ok := <-w.running:
				if !ok {
					break loop
				}
			default:
				// restarting process
				time.Sleep(time.Second)
			}
		}

		close(w.output)
		w.stopWaiting.Done()
	}()

	return w, nil
}

func (w *monitorWriteWidget) Read() ([]byte, error) {
	data, ok := <-w.output
	if !ok {
		return nil, io.EOF
	}

	return json.Marshal(data)
}

func (w *monitorWriteWidget) Write(data []byte) error {
	if w.connectWrite {
		var inputMessage executor.MonitorWriteInputMessage
		if err := json.Unmarshal(data, &inputMessage); err != nil {
			return err
		}

		w.input <- inputMessage
	}

	return nil
}

func (w *monitorWriteWidget) Close() {
	if w.connectWrite {
		close(w.input)
	}

	close(w.running)
	w.stopWaiting.Wait()
}

func (w *monitorWriteWidget) storeAndSendContents(contents []byte) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if bytes.Compare(contents, w.contents) == 0 {
		// nothing new to send
		return
	}

	w.contents = contents

	w.output <- executor.MonitorWriteOutputMessage{
		Contents: w.contents,
		Errors:   w.errors,
	}
}

func (w *monitorWriteWidget) storeAndSendError(err []byte) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if len(w.errors) > 4 {
		w.errors = append(w.errors[len(w.errors)-4:len(w.errors)], err)
	} else {
		w.errors = append(w.errors, err)
	}

	w.output <- executor.MonitorWriteOutputMessage{
		Contents: w.contents,
		Errors:   w.errors,
	}
}
