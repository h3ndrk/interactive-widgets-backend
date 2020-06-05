package fileio

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// type jsonError struct {
//     Type        string `json:"type"` // always "jsonError"
//     ErrorReason string `json:"errorReason"`
// }

// OpenError represents an error while creating a file
type OpenError struct {
	Type        string `json:"type"` // always "openError"
	Path        string `json:"path"`
	ErrorReason string `json:"errorReason"`
}

func (e *OpenError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":%s}", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// ReadAndDecodeError represents an error while reading and decoding Bas64 data
// from a file
type ReadAndDecodeError struct {
	Type        string `json:"type"` // always "readAndDecodeError"
	ErrorReason string `json:"errorReason"`
}

func (e *ReadAndDecodeError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":%s}", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// ReadFileToBase64 reads a file with given path and returns the Base64 encoded
// contents.
func ReadFileToBase64(pathToFile string) (string, error) {
	file, err := os.Open(pathToFile)
	if err != nil {
		return "", &OpenError{Type: "openError", Path: pathToFile, ErrorReason: err.Error()}
	}

	var encoded strings.Builder

	if err := func() error {
		encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
		defer encoder.Close()

		_, err = io.Copy(encoder, file)
		if err != nil {
			return &ReadAndDecodeError{Type: "readAndDecodeError", ErrorReason: err.Error()}
		}

		return nil
	}(); err != nil {
		return "", err
	}

	return encoded.String(), nil
}
