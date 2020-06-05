package fileio

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/fsnotify/fsnotify"
)

// type jsonError struct {
//     Type        string `json:"type"` // always "jsonError"
//     ErrorReason string `json:"errorReason"`
// }

// CreateWatcherError represents an error while creating a file watcher
type CreateWatcherError struct {
	Type        string `json:"type"` // always "createWatcherError"
	ErrorReason string `json:"errorReason"`
}

func (e *CreateWatcherError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":\"%s\"}\n", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// AddWatcherError represents an error while creating a file watcher
type AddWatcherError struct {
	Type        string `json:"type"` // always "addWatcherError"
	Path        string `json:"path"`
	ErrorReason string `json:"errorReason"`
}

func (e *AddWatcherError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":\"%s\"}\n", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// WatchError represents an error while watching a file
type WatchError struct {
	Type        string `json:"type"` // always "watchError"
	Path        string `json:"path"`
	ErrorReason string `json:"errorReason"`
}

func (e *WatchError) Error() string {
	marshalled, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("{\"type\":\"jsonError\",\"errorReason\":\"%s\"}\n", strconv.Quote(err.Error()))
	}

	return string(marshalled)
}

// WaitForEvent watches a path via inotify and returns whether the watched file
// or one of the parent directories changed.
func WaitForEvent(pathToWatch string, done <-chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return &CreateWatcherError{Type: "createWatcherError", ErrorReason: err.Error()}
	}
	defer watcher.Close()

	if err := watcher.Add(pathToWatch); err != nil {
		return &AddWatcherError{Type: "addWatcherError", Path: pathToWatch, ErrorReason: err.Error()}
	}
	var parentPaths []string
	currentPath := pathToWatch
	for filepath.Dir(currentPath) != currentPath {
		currentPath = filepath.Dir(currentPath)
		parentPaths = append(parentPaths, currentPath)
	}
	for _, parentPath := range parentPaths {
		if err := watcher.Add(parentPath); err != nil {
			return &AddWatcherError{Type: "addWatcherError", Path: parentPath, ErrorReason: err.Error()}
		}
	}

	for {
		select {
		case _, ok := <-done:
			if !ok {
				return nil
			}
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if event.Name == pathToWatch && event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				return nil
			}

			for _, parentPath := range parentPaths {
				if event.Name == parentPath && event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
					return nil
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}

			return &WatchError{Type: "watchError", Path: pathToWatch, ErrorReason: err.Error()}
		}
	}
}
