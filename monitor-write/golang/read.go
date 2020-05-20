package main

import (
	"encoding/base64"
	"io"
	"os"
)

func readFileAndOutputBase64(pathToFile string) error {
	file, err := os.Open("/home/hendrik/Documents/containerized-playground/b/b/c/test.txt")
	if err != nil {
		return err
	}

	if err := func() error {
		encoder := base64.NewEncoder(base64.StdEncoding, os.Stdout)
		defer encoder.Close()

		_, err = io.Copy(encoder, file)
		if err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return nil
	}
	os.Stdout.WriteString("\n")

	return nil
}
