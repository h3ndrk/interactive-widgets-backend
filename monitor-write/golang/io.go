package main

import (
	"encoding/base64"
	"io"
	"os"
	"strings"
)

func readFileToBase64(pathToFile string) (string, error) {
	file, err := os.Open(pathToFile)
	if err != nil {
		return "", err
	}

	var encoded strings.Builder

	if err := func() error {
		encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
		defer encoder.Close()

		_, err = io.Copy(encoder, file)
		if err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return "", nil
	}

	return encoded.String(), nil
}

func writeFileFromBase64(pathToFile string, encoded string) error {
	file, err := os.Create(pathToFile)
	if err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}

	_, err = file.Write(decoded)
	if err != nil {
		return err
	}

	return nil
}
