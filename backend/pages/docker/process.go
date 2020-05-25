package docker

import (
	"bufio"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

type OutputStream string

const StdoutStream OutputStream = "StdoutStream"
const StderrStream OutputStream = "StderrStream"

type Data struct {
	Origin OutputStream
	Bytes  []byte
}

type LongRunningProcess struct {
	running     chan struct{}
	OutputData  <-chan Data
	stopWaiting *sync.WaitGroup
}

func NewLongRunningProcess(arguments []string, stdin <-chan []byte) (*LongRunningProcess, error) {
	if len(arguments) <= 1 {
		return nil, errors.New("too few arguments")
	}

	running := make(chan struct{})
	outputData := make(chan Data)
	var stopWaiting sync.WaitGroup

	go func() {
		stopWaiting.Add(1)
	loop:
		for {
			done := make(chan struct{})
			process := exec.Command(arguments[0], arguments[1:]...)

			stdinWriter, err := process.StdinPipe()
			if err != nil {
				log.Print(err)
				time.Sleep(time.Second)
				continue
			}

			stdoutPipe, err := process.StdoutPipe()
			if err != nil {
				log.Print(err)
				time.Sleep(time.Second)
				continue
			}
			stdoutScanner := bufio.NewScanner(stdoutPipe)

			stderrPipe, err := process.StderrPipe()
			if err != nil {
				log.Print(err)
				time.Sleep(time.Second)
				continue
			}
			stderrScanner := bufio.NewScanner(stderrPipe)

			err = process.Start()
			if err != nil {
				log.Print(err)
				time.Sleep(time.Second)
				continue
			}

			go func() {
				defer stdinWriter.Close()
				for {
					select {
					case data, ok := <-stdin:
						if !ok {
							return
						}
						_, err := stdinWriter.Write(append(data, "\n"...))
						if err != nil {
							log.Print(err)
							continue
						}
					case <-done:
						return
					}
				}
			}()
			go func() {
				for stdoutScanner.Scan() {
					outputData <- Data{
						Origin: StdoutStream,
						Bytes:  stdoutScanner.Bytes(),
					}
				}
			}()
			go func() {
				for stderrScanner.Scan() {
					outputData <- Data{
						Origin: StderrStream,
						Bytes:  stderrScanner.Bytes(),
					}
				}
			}()
			go func() {
				select {
				case <-done:
				case _, ok := <-running:
					if !ok {
						if err := process.Process.Signal(syscall.SIGTERM); err != nil {
							log.Print(err)
						}
					}
				}
			}()

			process.Wait()
			close(done)

			select {
			case _, ok := <-running:
				if !ok {
					break loop
				}
			default:
				log.Print("Restarting process ...")
				time.Sleep(time.Second)
			}
		}
		close(outputData)
		stopWaiting.Done()
	}()

	return &LongRunningProcess{
		running:     running,
		OutputData:  outputData,
		stopWaiting: &stopWaiting,
	}, nil
}

func (l LongRunningProcess) Stop() {
	close(l.running)
	l.stopWaiting.Wait()
}

type ShortRunningProcess struct {
	OutputData  <-chan Data
	stopWaiting *sync.WaitGroup
}

func NewShortRunningProcess(arguments []string) (*ShortRunningProcess, error) {
	if len(arguments) <= 1 {
		return nil, errors.New("too few arguments")
	}

	outputData := make(chan Data)
	var stopWaiting sync.WaitGroup

	stopWaiting.Add(1)
	process := exec.Command(arguments[0], arguments[1:]...)

	stdoutPipe, err := process.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdoutScanner := bufio.NewScanner(stdoutPipe)

	stderrPipe, err := process.StderrPipe()
	if err != nil {
		return nil, err
	}
	stderrScanner := bufio.NewScanner(stderrPipe)

	err = process.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		for stdoutScanner.Scan() {
			outputData <- Data{
				Origin: StdoutStream,
				Bytes:  stdoutScanner.Bytes(),
			}
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			outputData <- Data{
				Origin: StderrStream,
				Bytes:  stderrScanner.Bytes(),
			}
		}
	}()

	go func(stopWaiting *sync.WaitGroup) {
		process.Wait()
		close(outputData)
		stopWaiting.Done()
	}(&stopWaiting)

	return &ShortRunningProcess{
		OutputData:  outputData,
		stopWaiting: &stopWaiting,
	}, nil
}

func (s ShortRunningProcess) Wait() {
	s.stopWaiting.Wait()
}
