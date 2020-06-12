package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/h3ndrk/inter-md/internal/fileio"
)

// jsonError represents an error while marshalling/unmarshalling JSON data
// type jsonError struct {
//     Type        string `json:"type"` // always "jsonError"
//     ErrorReason string `json:"errorReason"`
// }

// argumentError represents an error when sanitizing input arguments
type argumentError struct {
	Type          string   `json:"type"` // always "argumentError"
	ExpectedCount int      `json:"expectedCount"`
	GotCount      int      `json:"gotCount"`
	GotArguments  []string `json:"gotArguments"`
}

// stdinReadError represents an error while reading from stdin
type stdinReadError struct {
	Type        string `json:"type"` // always "stdinReadError"
	ErrorReason string `json:"errorReason"`
}

// removalError represents an error while removing the watched file
type removalError struct {
	Type        string `json:"type"` // always "removalError"
	Path        string `json:"path"`
	ErrorReason string `json:"errorReason"`
}

// contents represents a request for writing the given contents to the file
type contents struct {
	Type     string `json:"type"`     // always "contents"
	Contents string `json:"contents"` // as Base64
}

// removal represents a request to delete the watched file
type removal struct {
	Type string `json:"type"` // always "removal"
}

func main() {
	stdoutEncoder := json.NewEncoder(os.Stdout)
	if len(os.Args) != 2 {
		if err := stdoutEncoder.Encode(&argumentError{
			Type:          "argumentError",
			ExpectedCount: 2,
			GotCount:      len(os.Args),
			GotArguments:  os.Args,
		}); err != nil {
			fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":%s}\n", strconv.Quote(err.Error()))
		}
		return
	}
	pathToWatch := os.Args[1]

	done := make(chan struct{}, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		close(done)
	}()

	go func() {
		stderrEncoder := json.NewEncoder(os.Stderr)
		scanner := bufio.NewScanner(os.Stdin)
		buffer := make([]byte, 0, 64*1024)
		scanner.Buffer(buffer, 1024*1024)

	stdinLoop:
		for scanner.Scan() {
			var typeMessage struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(scanner.Bytes(), &typeMessage); err != nil {
				fmt.Fprintf(os.Stderr, "{\"type\":\"jsonError\",\"errorReason\":%s}\n", strconv.Quote(err.Error()))
				continue
			}

			switch typeMessage.Type {
			case "contents":
				var contentsMessage contents
				if err := json.Unmarshal(scanner.Bytes(), &contentsMessage); err != nil {
					fmt.Fprintf(os.Stderr, "{\"type\":\"jsonError\",\"errorReason\":%s}\n", strconv.Quote(err.Error()))
					continue stdinLoop
				}

				if err := fileio.WriteFileFromBase64(pathToWatch, contentsMessage.Contents); err != nil {
					fmt.Println(err.Error())
					continue stdinLoop
				}
			case "removal":
				if err := os.Remove(pathToWatch); err != nil {
					if err := stderrEncoder.Encode(&removalError{
						Type:        "removalError",
						Path:        pathToWatch,
						ErrorReason: err.Error(),
					}); err != nil {
						fmt.Fprintf(os.Stderr, "{\"type\":\"jsonError\",\"errorReason\":%s}\n", strconv.Quote(err.Error()))
					}
					continue stdinLoop
				}
			}
		}
		if err := scanner.Err(); err != nil {
			if err := stderrEncoder.Encode(&stdinReadError{
				Type:        "stdinReadError",
				ErrorReason: err.Error(),
			}); err != nil {
				fmt.Fprintf(os.Stderr, "{\"type\":\"jsonError\",\"errorReason\":%s}\n", strconv.Quote(err.Error()))
			}
		}
	}()

	lastEncoded := ""
stdoutLoop:
	for {
		encoded, err := fileio.ReadFileToBase64(pathToWatch)
		if err != nil {
			if err.Error() != lastEncoded {
				fmt.Println(err.Error())
				lastEncoded = err.Error()
			}
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-timer.C:
			case <-done:
				if !timer.Stop() {
					<-timer.C
				}
				break stdoutLoop
			}
			continue
		}

		// encoded will be empty or containing the base64 string
		if encoded != lastEncoded {
			if err := stdoutEncoder.Encode(&contents{
				Type:     "contents",
				Contents: encoded,
			}); err != nil {
				fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":%s}\n", strconv.Quote(err.Error()))
			}
			lastEncoded = encoded
		}

		if err := fileio.WaitForEvent(pathToWatch, done); err != nil {
			if err.Error() != lastEncoded {
				fmt.Println(err.Error())
				lastEncoded = err.Error()
			}
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-timer.C:
			case <-done:
				if !timer.Stop() {
					<-timer.C
				}
				break stdoutLoop
			}
			continue
		}

		select {
		case _, ok := <-done:
			if !ok {
				break stdoutLoop
			}
		default:
		}
	}
}
