package fileio

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// jsonError represents an error while marshalling/unmarshalling JSON data
// type jsonError struct {
//     Type        string `json:"type"` // always "jsonError"
//     ErrorReason string `json:"errorReason"`
// }

// createError represents an error while creating a file
type createError struct {
	Type        string `json:"type"` // always "createError"
	Path        string `json:"path"`
	ErrorReason string `json:"errorReason"`
}

func (e *createError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":%s}", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// decodeError represents an error while decoding Base64 data
type decodeError struct {
	Type        string `json:"type"` // always "decodeError"
	ErrorReason string `json:"errorReason"`
}

func (e *decodeError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":%s}", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// writeError represents an error while writing Base64 data to a file
type writeError struct {
	Type        string `json:"type"` // always "writeError"
	ErrorReason string `json:"errorReason"`
}

func (e *writeError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":%s}", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// WriteFileFromBase64 writes a file with given path and given Base64 encoded
// contents.
func WriteFileFromBase64(pathToFile string, encoded string) error {
	file, err := os.Create(pathToFile)
	if err != nil {
		return &createError{Type: "createError", Path: pathToFile, ErrorReason: err.Error()}
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return &decodeError{Type: "decodeError", ErrorReason: err.Error()}
	}

	_, err = file.Write(decoded)
	if err != nil {
		return &writeError{Type: "writeError", ErrorReason: err.Error()}
	}

	return nil
}
