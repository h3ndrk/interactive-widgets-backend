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

// jsonError represents an error while marshalling/unmarshalling JSON data
// type jsonError struct {
//     Type        string `json:"type"` // always "jsonError"
//     ErrorReason string `json:"errorReason"`
// }

// openError represents an error while creating a file
type openError struct {
	Type        string `json:"type"` // always "openError"
	Path        string `json:"path"`
	ErrorReason string `json:"errorReason"`
}

func (e *openError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":%s}", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// readAndDecodeError represents an error while reading and decoding Bas64 data
// from a file
type readAndDecodeError struct {
	Type        string `json:"type"` // always "readAndDecodeError"
	ErrorReason string `json:"errorReason"`
}

func (e *readAndDecodeError) Error() string {
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
		return "", &openError{Type: "openError", Path: pathToFile, ErrorReason: err.Error()}
	}

	var encoded strings.Builder

	if err := func() error {
		encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
		defer encoder.Close()

		_, err = io.Copy(encoder, file)
		if err != nil {
			return &readAndDecodeError{Type: "readAndDecodeError", ErrorReason: err.Error()}
		}

		return nil
	}(); err != nil {
		return "", err
	}

	return encoded.String(), nil
}
