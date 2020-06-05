package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/h3ndrk/containerized-playground/internal/fileio"
	"github.com/pkg/errors"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Expected 2 parameters, got", len(os.Args), "parameters:", os.Args)
		return
	}
	pathToWatch := os.Args[1]

	done := make(chan struct{}, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		fmt.Fprintln(os.Stderr, "Got ", <-signals)
		close(done)
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			err := fileio.WriteFileFromBase64(pathToWatch, scanner.Text())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				break
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, "Error while reading stdin"))
		}
	}()

	lastEncoded := ""
	didOutputAtLeastOnce := false
loop:
	for {
		encoded, err := fileio.ReadFileToBase64(pathToWatch)
		// encoded will be empty or containing the base64 string
		if encoded != lastEncoded {
			didOutputAtLeastOnce = true
			fmt.Println(encoded)
			lastEncoded = encoded
		} else if !didOutputAtLeastOnce && encoded == "" {
			didOutputAtLeastOnce = true
			fmt.Println(encoded)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-timer.C:
			case <-done:
				if !timer.Stop() {
					<-timer.C
				}
				break loop
			}
			continue
		}

		if err := fileio.WaitForEvent(pathToWatch, done); err != nil {
			fmt.Fprintln(os.Stderr, err)
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-timer.C:
			case <-done:
				if !timer.Stop() {
					<-timer.C
				}
				break loop
			}
			continue
		}

		select {
		case _, ok := <-done:
			if !ok {
				break loop
			}
		default:
		}
	}
}
