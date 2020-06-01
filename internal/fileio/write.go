package fileio

import (
	"encoding/base64"
	"os"

	"github.com/pkg/errors"
)

// WriteFileFromBase64 writes a file with given path and given Base64 encoded
// contents.
func WriteFileFromBase64(pathToFile string, encoded string) error {
	file, err := os.Create(pathToFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to open \"%s\" for writing", pathToFile)
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return errors.Wrap(err, "Failed to decode base64 data")
	}

	_, err = file.Write(decoded)
	if err != nil {
		return errors.Wrapf(err, "Failed to write base64 data to \"%s\"", pathToFile)
	}

	return nil
}
