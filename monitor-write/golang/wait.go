package main

import (
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

func waitForEvent(pathToWatch string, done <-chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrapf(err, "failed to create new watcher")
	}
	defer watcher.Close()

	if err := watcher.Add(pathToWatch); err != nil {
		return errors.Wrapf(err, "failed to add watched child path %s", pathToWatch)
	}
	var parentPaths []string
	currentPath := pathToWatch
	for filepath.Dir(currentPath) != currentPath {
		currentPath = filepath.Dir(currentPath)
		parentPaths = append(parentPaths, currentPath)
	}
	for _, parentPath := range parentPaths {
		if err := watcher.Add(parentPath); err != nil {
			return errors.Wrapf(err, "failed to add watched parent path %s", parentPath)
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
			return errors.Wrap(err, "error while watching")
		}
	}
}
