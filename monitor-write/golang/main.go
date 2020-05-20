package main

import (
	"log"
)

func main() {
	done := make(chan struct{})
	pathToWatch := "/home/hendrik/Documents/containerized-playground/b/b/c/test.txt"
	if err := readFileAndOutputBase64(pathToWatch); err != nil {
		log.Fatal(err)
	}
	if err := waitForEvent(pathToWatch, done); err != nil {
		log.Fatal(err)
	}
}
