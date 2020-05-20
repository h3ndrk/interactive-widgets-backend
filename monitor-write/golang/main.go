package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	done := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		fmt.Fprintln(os.Stderr, "Received signal:", <-signals)
		close(done)
	}()
	pathToWatch := "/home/hendrik/Documents/containerized-playground/b/b/c/test.txt"
	lastEncoded := ""
	// we need to first output an empty line because otherwise we don't output an empty line when the read results in an error
	fmt.Println(lastEncoded)
	for {
		encoded, err := readFileToBase64(pathToWatch)
		// encoded will be empty or containing the base64 string
		if encoded != lastEncoded {
			fmt.Println(encoded)
			lastEncoded = encoded
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
			}
			if _, ok := <-done; !ok {
				break
			}
			continue
		}
		if err := waitForEvent(pathToWatch, done); err != nil {
			fmt.Fprintln(os.Stderr, err)
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-timer.C:
			case <-done:
				if !timer.Stop() {
					<-timer.C
				}
			}
			if _, ok := <-done; !ok {
				break
			}
			continue
		}
	}
}
