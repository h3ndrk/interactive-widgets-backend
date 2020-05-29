package docker

import (
	"sync"

	"github.com/h3ndrk/containerized-playground/internal/executor"
	"github.com/h3ndrk/containerized-playground/internal/id"
	"github.com/h3ndrk/containerized-playground/internal/parser"
	"github.com/pkg/errors"
)

type widgetStream interface {
	Read() ([]byte, error)
	Write([]byte) error
	Close()
}

type Executor struct {
	pages        []parser.Page
	widgetsMutex sync.Mutex
	widgets      map[id.WidgetID]widgetStream
}

func NewExecutor(pages []parser.Page) (executor.Executor, error) {
	return &Executor{
		pages: pages,
	}, nil
}

func (e *Executor) pageFromPageURL(pageURL id.PageURL) *parser.Page {
	for i, page := range e.pages {
		if page.URL == pageURL {
			return &e.pages[i]
		}
	}

	return nil
}

func (e *Executor) StartPage(pageID id.PageID) error {
	// TODO: define error types
	pageURL, roomID, err := id.PageURLAndRoomIDFromPageID(pageID)
	if err != nil {
		return err
	}

	page := e.pageFromPageURL(pageURL)
	if page == nil {
		return errors.Errorf("No page with URL %s", pageURL)
	}

	// TODO: create volume

	e.widgetsMutex.Lock()
	defer e.widgetsMutex.Unlock()

	var temporaryWidgets map[id.WidgetID]widgetStream
	defer func() {
		// in case of error: close all temporary widgets
		var closeWaiting sync.WaitGroup
		closeWaiting.Add(len(temporaryWidgets))
		for widgetID, widget := range temporaryWidgets {
			go func(widget widgetStream, closeWaiting *sync.WaitGroup) {
				widget.Close()
				closeWaiting.Done()
			}()
		}
		closeWaiting.Wait()
	}()

	for widgetIndex, widget := range page.Widgets {
		widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
		if err != nil {
			return err
		}

		switch widget := widget.(type) {
		case parser.TextWidget:
			textWidget, err := newMonitorWriteWidget(widgetID, widget.File, false)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = textWidget
		case parser.ImageWidget:
			imageWidget, err := newMonitorWriteWidget(widgetID, widget.File, false)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = imageWidget
		case parser.ButtonWidget:
			buttonWidget, err := newButtonWidget(widgetID, widget)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = buttonWidget
		case parser.EditorWidget:
			editorWidget, err := newMonitorWriteWidget(widgetID, widget.File, true)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = editorWidget
		case parser.TerminalWidget:
			terminalWidget, err := newTerminalWidget(widgetID, widget)
			if err != nil {
				return err
			}

			temporaryWidgets[widgetID] = terminalWidget
		default:
			panic("Not implemented")
		}
	}

	// if reached here: copy widgets into executor and remove them from temporary widgets
	for widgetID, widget := range temporaryWidgets {
		e.widgets[widgetID] = widget
		delete(temporaryWidgets, widgetID)
	}

	return nil
}

func (e *Executor) StopPage(pageID id.PageID) error {
	pageURL, roomID, err := id.PageURLAndRoomIDFromPageID(pageID)
	if err != nil {
		return err
	}

	page := e.pageFromPageURL(pageURL)
	if page == nil {
		return errors.Errorf("No page with URL %s", pageURL)
	}

	e.widgetsMutex.Lock()
	defer e.widgetsMutex.Unlock()

	// close all widgets and remove them
	var closeWaiting sync.WaitGroup
	closeWaiting.Add(len(widgets))
	for widgetIndex := range page.Widgets {
		widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
		if err != nil {
			return err
		}

		go func(widget widgetStream, closeWaiting *sync.WaitGroup) {
			widget.Close()
			closeWaiting.Done()
		}()
	}
	closeWaiting.Wait()
	for widgetIndex := range page.Widgets {
		widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
		if err != nil {
			return err
		}

		delete(e.widgets, widgetID)
	}

	// TODO: remove volume

	return nil
}

func (e *Executor) Read(widgetID id.WidgetID) ([]byte, error) {
	widget, ok := e.widgets[widgetID]
	if !ok {
		return nil, errors.Errorf("No widget with ID %s", widgetID)
	}

	return widget.Read()
}

func (e *Executor) Write(widgetID id.WidgetID, data []byte) error {
	widget, ok := e.widgets[widgetID]
	if !ok {
		return errors.Errorf("No widget with ID %s", widgetID)
	}

	return widget.Write(data)
}
