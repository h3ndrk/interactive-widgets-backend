package fileio

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// type jsonError struct {
//     Type        string `json:"type"` // always "jsonError"
//     ErrorReason string `json:"errorReason"`
// }

// CreateError represents an error while creating a file
type CreateError struct {
	Type        string `json:"type"` // always "createError"
	Path        string `json:"path"`
	ErrorReason string `json:"errorReason"`
}

func (e *CreateError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":\"%s\"}\n", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// DecodeError represents an error while decoding Base64 data
type DecodeError struct {
	Type        string `json:"type"` // always "decodeError"
	ErrorReason string `json:"errorReason"`
}

func (e *DecodeError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":\"%s\"}\n", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// WriteError represents an error while writing Base64 data to a file
type WriteError struct {
	Type        string `json:"type"` // always "writeError"
	ErrorReason string `json:"errorReason"`
}

func (e *WriteError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":\"%s\"}\n", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// WriteFileFromBase64 writes a file with given path and given Base64 encoded
// contents.
func WriteFileFromBase64(pathToFile string, encoded string) error {
	file, err := os.Create(pathToFile)
	if err != nil {
		return &CreateError{Type: "createError", Path: pathToFile, ErrorReason: err.Error()}
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return &DecodeError{Type: "decodeError", ErrorReason: err.Error()}
	}

	_, err = file.Write(decoded)
	if err != nil {
		return &WriteError{Type: "writeError", ErrorReason: err.Error()}
	}

	return nil
}
