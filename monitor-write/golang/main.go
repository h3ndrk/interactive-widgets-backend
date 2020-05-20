package main

import (
	"log"
)

func main() {
	done := make(chan struct{})
	if err := waitForEvent("/home/hendrik/Documents/containerized-playground/b/b/c/test.txt", done); err != nil {
		log.Fatal(err)
	}
}
