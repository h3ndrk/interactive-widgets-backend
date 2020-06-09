package docker

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/h3ndrk/containerized-playground/internal/executor"
	"github.com/h3ndrk/containerized-playground/internal/id"
	"github.com/h3ndrk/containerized-playground/internal/parser"
	"github.com/pkg/errors"
)

type widgetStream interface {
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
	GetCurrentState() ([]byte, error)
}

// Executor implements the executor.Executor interface.
type Executor struct {
	pages        []parser.Page
	widgetsMutex sync.Mutex
	widgets      map[id.WidgetID]widgetStream
}

// NewExecutor creates a new executor from pages.
func NewExecutor(pages []parser.Page) (executor.Executor, error) {
	return &Executor{
		pages:   pages,
		widgets: map[id.WidgetID]widgetStream{},
	}, nil
}

// StartPage creates a docker volume and starts all widget containers.
func (e *Executor) StartPage(pageID id.PageID) error {
	pageURL, roomID, err := id.PageURLAndRoomIDFromPageID(pageID)
	if err != nil {
		return errors.Wrapf(err, "Failed to decode page ID \"%s\"", pageID)
	}

	page := parser.PageFromPageURL(e.pages, pageURL)
	if page == nil {
		return errors.Errorf("No page with URL \"%s\"", pageURL)
	}

	if !page.IsInteractive {
		return errors.Errorf("Page \"%s\" is not interactive", pageURL)
	}

	// create volume
	volumeName := fmt.Sprintf("containerized-playground-%s", id.EncodePageID(pageID))

	process := exec.Command("docker", "volume", "create", volumeName)
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr

	err = process.Start()
	if err != nil {
		return errors.Wrapf(err, "Failed to create volume for page \"%s\"", pageID)
	}

	process.Wait()

	// start widgets
	e.widgetsMutex.Lock()
	defer e.widgetsMutex.Unlock()

	temporaryWidgets := map[id.WidgetID]widgetStream{}
	defer func() {
		// in case of error: close all temporary widgets
		var closeWaiting sync.WaitGroup
		closeWaiting.Add(len(temporaryWidgets))
		for _, widget := range temporaryWidgets {
			go func(widget widgetStream, closeWaiting *sync.WaitGroup) {
				widget.Close()
				closeWaiting.Done()
			}(widget, &closeWaiting)
		}
		closeWaiting.Wait()
	}()

	for widgetIndex, widget := range page.Widgets {
		if !widget.IsInteractive() {
			continue
		}

		widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
		if err != nil {
			return err
		}

		switch widget := widget.(type) {
		case *parser.TextWidget:
			textWidget, err := newMonitorWriteWidget(widgetID, widget.File, false)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = textWidget
		case *parser.ImageWidget:
			imageWidget, err := newMonitorWriteWidget(widgetID, widget.File, false)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = imageWidget
		case *parser.ButtonWidget:
			buttonWidget, err := newButtonWidget(widgetID, widget)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = buttonWidget
		case *parser.EditorWidget:
			editorWidget, err := newMonitorWriteWidget(widgetID, widget.File, true)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = editorWidget
		case *parser.TerminalWidget:
			terminalWidget, err := newTerminalWidget(widgetID, widget)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = terminalWidget
		default:
			return errors.Wrapf(err, "Interactive widget not implemented: %T", widget)
		}
	}

	// if reached here: copy widgets into executor and remove them from temporary widgets
	for widgetID, widget := range temporaryWidgets {
		e.widgets[widgetID] = widget
		delete(temporaryWidgets, widgetID)
	}

	return nil
}

// StopPage removes the docker volume and stops all widget containers.
func (e *Executor) StopPage(pageID id.PageID) error {
	pageURL, roomID, err := id.PageURLAndRoomIDFromPageID(pageID)
	if err != nil {
		return errors.Wrapf(err, "Failed to decode page ID \"%s\"", pageID)
	}

	page := parser.PageFromPageURL(e.pages, pageURL)
	if page == nil {
		return errors.Errorf("No page with URL \"%s\"", pageURL)
	}

	if !page.IsInteractive {
		return errors.Errorf("Page \"%s\" is not interactive", pageURL)
	}

	// close all widgets
	var closeWaiting sync.WaitGroup
	for widgetIndex, widget := range page.Widgets {
		if !widget.IsInteractive() {
			continue
		}

		widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
		if err != nil {
			return err
		}

		closeWaiting.Add(1)
		go func(widget widgetStream, closeWaiting *sync.WaitGroup) {
			widget.Close()
			closeWaiting.Done()
		}(e.widgets[widgetID], &closeWaiting)
	}
	closeWaiting.Wait()

	e.widgetsMutex.Lock()
	defer e.widgetsMutex.Unlock()

	// remove all widgets
	for widgetIndex, widget := range page.Widgets {
		if !widget.IsInteractive() {
			continue
		}

		widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
		if err != nil {
			return err
		}

		delete(e.widgets, widgetID)
	}

	// remove volume
	volumeName := fmt.Sprintf("containerized-playground-%s", id.EncodePageID(pageID))

	process := exec.Command("docker", "volume", "rm", volumeName)
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr

	err = process.Start()
	if err != nil {
		return errors.Wrapf(err, "Failed to remove volume for page \"%s\"", pageID)
	}

	process.Wait()

	return nil
}

// Read returns data from the widget with given widget ID.
func (e *Executor) Read(widgetID id.WidgetID) ([]byte, error) {
	e.widgetsMutex.Lock()
	widget, ok := e.widgets[widgetID]
	e.widgetsMutex.Unlock()
	if !ok {
		return nil, errors.Wrapf(errors.New("No widget with ID existing"), "Failed to read from widget \"%s\"", widgetID)
	}

	return widget.Read()
}

// Write sends given data to the widget with given widget ID.
func (e *Executor) Write(widgetID id.WidgetID, data []byte) error {
	e.widgetsMutex.Lock()
	widget, ok := e.widgets[widgetID]
	e.widgetsMutex.Unlock()
	if !ok {
		return errors.Wrapf(errors.New("No widget with ID existing"), "Failed to write to widget \"%s\"", widgetID)
	}

	return widget.Write(data)
}

// GetCurrentState retrieves the current state from the widget with given
// widget ID.
func (e *Executor) GetCurrentState(widgetID id.WidgetID) ([]byte, error) {
	e.widgetsMutex.Lock()
	widget, ok := e.widgets[widgetID]
	e.widgetsMutex.Unlock()
	if !ok {
		return nil, errors.Wrapf(errors.New("No widget with ID existing"), "Failed to get current state from widget \"%s\"", widgetID)
	}

	return widget.GetCurrentState()
}

// BuildImages builds all pages images and tags them s.t. this executor is able
// to use them when executed as backend executor.
func (e *Executor) BuildImages() error {
	for _, page := range e.pages {
		if page.IsInteractive {
			log.Printf("Building docker image for interactive page \"%s\" ...", page.URL)

			imageName := fmt.Sprintf("containerized-playground-%s", id.EncodePageURL(page.URL))

			process := exec.Command("docker", "build", "--pull", "--tag", imageName, page.BasePath)
			process.Stdout = os.Stdout
			process.Stderr = os.Stderr

			if err := process.Start(); err != nil {
				return errors.Wrapf(err, "Failed to build image for page \"%s\"", page.URL)
			}

			if err := process.Wait(); err != nil {
				return errors.Wrapf(err, "Failed to build image for page \"%s\"", page.URL)
			}
		}
	}

	return nil
}
