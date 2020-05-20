package main

import (
	"encoding/base64"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func readFileToBase64(pathToFile string) (string, error) {
	file, err := os.Open(pathToFile)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open %s for reading", pathToFile)
	}

	var encoded strings.Builder

	if err := func() error {
		encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
		defer encoder.Close()

		_, err = io.Copy(encoder, file)
		if err != nil {
			return errors.Wrapf(err, "failed to read base64 data from %s", pathToFile)
		}

		return nil
	}(); err != nil {
		return "", err
	}

	return encoded.String(), nil
}

func writeFileFromBase64(pathToFile string, encoded string) error {
	file, err := os.Create(pathToFile)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s for writing", pathToFile)
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return errors.Wrap(err, "failed to decode base64 data")
	}

	_, err = file.Write(decoded)
	if err != nil {
		return errors.Wrapf(err, "failed to write base64 data to %s", pathToFile)
	}

	return nil
}
