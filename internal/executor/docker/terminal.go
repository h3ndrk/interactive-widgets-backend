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

type terminalWidget struct {
	running     chan struct{}
	stopWaiting *sync.WaitGroup
	input       chan executor.TerminalInputMessage
	output      chan executor.TerminalOutputMessage

	mutex    sync.Mutex
	contents []byte
	errors   [][]byte
}

func newTerminalWidget(widgetID id.WidgetID, widget parser.TerminalWidget) (widgetStream, error) {
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
		running:     make(chan struct{}),
		stopWaiting: &sync.WaitGroup{},
		output:      make(chan executor.TerminalOutputMessage),
		input:       make(chan executor.TerminalInputMessage),
	}

	go func() {
		w.stopWaiting.Add(1)
	loop:
		for {
			done := make(chan struct{})
			process := exec.Command("docker", "run", "--rm", "--interactive", "--tty", "--name", containerName, "--network=none", "--mount", fmt.Sprintf("src=%s,dst=/data", volumeName), "--workdir", widget.WorkingDirectory, imageName, "/bin/bash")
			pseudoTerminal, err := pty.Start(process)
			if err != nil {
				w.output <- executor.TerminalOutputMessage{
					Data: []byte(err.Error() + "\n"),
				}
				time.Sleep(time.Second)
				continue
			}

			go func() {
				defer pseudoTerminal.Close()

				for {
					select {
					case message, ok := <-w.input:
						if !ok {
							return
						}

						_, err := pseudoTerminal.Write(message.Data)
						if err != nil {
							w.output <- executor.TerminalOutputMessage{
								Data: []byte(err.Error() + "\n"),
							}
						}
					case <-done:
						return
					}
				}
			}()

			go func() {
				chunk := make([]byte, 4096)

				for {
					n, err := pseudoTerminal.Read(chunk)
					if err != nil {
						if err == io.EOF || errors.Is(err, os.ErrClosed) {
							return
						}

						w.output <- executor.TerminalOutputMessage{
							Data: []byte(err.Error() + "\n"),
						}
						time.Sleep(time.Second)
						continue
					}

					w.output <- executor.TerminalOutputMessage{
						Data: chunk[:n],
					}
				}
			}()

			go func() {
				select {
				case <-done:
				case _, ok := <-w.running:
					if !ok {
						if err := process.Process.Signal(syscall.SIGTERM); err != nil {
							w.output <- executor.TerminalOutputMessage{
								Data: []byte(err.Error() + "\n"),
							}
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

func (w *terminalWidget) Read() ([]byte, error) {
	data, ok := <-w.output
	if !ok {
		return nil, io.EOF
	}

	return json.Marshal(data)
}

func (w *terminalWidget) Write(data []byte) error {
	var inputMessage executor.TerminalInputMessage
	if err := json.Unmarshal(data, &inputMessage); err != nil {
		return err
	}

	w.input <- inputMessage

	return nil
}

func (w *terminalWidget) Close() {
	close(w.running)
	w.stopWaiting.Wait()
}
