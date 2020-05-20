package main

import (
	"log"
)

func main() {
	done := make(chan struct{})
	pathToWatch := "/home/hendrik/Documents/containerized-playground/b/b/c/test.txt"
	encoded, err := readFileToBase64(pathToWatch)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(encoded)
	if err := waitForEvent(pathToWatch, done); err != nil {
		log.Fatal(err)
	}
}
