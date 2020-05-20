package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
		fmt.Fprintln(os.Stderr, "Received signal:", <-signals)
		close(done)
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			err := writeFileFromBase64(pathToWatch, scanner.Text())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				break
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, "error while reading stdin"))
		}
		wg.Done()
	}(&wg)
	lastEncoded := ""
	didOutputAtLeastOnce := false
main:
	for {
		encoded, err := readFileToBase64(pathToWatch)
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
				break main
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
				break main
			}
			continue
		}
		select {
		case _, ok := <-done:
			if !ok {
				break main
			}
		default:
		}
	}
	wg.Wait()
}
