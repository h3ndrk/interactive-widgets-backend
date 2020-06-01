package fileio

import (
	"encoding/base64"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// ReadFileToBase64 reads a file with given path and returns the Base64 encoded
// contents.
func ReadFileToBase64(pathToFile string) (string, error) {
	file, err := os.Open(pathToFile)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to open \"%s\" for reading", pathToFile)
	}

	var encoded strings.Builder

	if err := func() error {
		encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
		defer encoder.Close()

		_, err = io.Copy(encoder, file)
		if err != nil {
			return errors.Wrapf(err, "Failed to read base64 data from \"%s\"", pathToFile)
		}

		return nil
	}(); err != nil {
		return "", err
	}

	return encoded.String(), nil
}
