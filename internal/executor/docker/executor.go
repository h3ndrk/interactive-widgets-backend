package docker

import (
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
	pages   []parser.Page
	widgets map[id.WidgetID]widgetStream
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

	for widgetIndex, widget := range page.Widgets {
		widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
		if err != nil {
			return err
		}

		switch widget := widget.(type) {
		case parser.TextWidget:
			textWidget, err := newMonitorWriteWidget(widgetID, widget, false)
			if err != nil {
				return err
			}

			e.widgets[widgetID] = textWidget
		case parser.ImageWidget:
			imageWidget, err := newMonitorWriteWidget(widgetID, widget, false)
			if err != nil {
				return err
			}

			e.widgets[widgetID] = imageWidget
		case parser.EditorWidget:
			editorWidget, err := newMonitorWriteWidget(widgetID, widget, true)
			if err != nil {
				return err
			}

			e.widgets[widgetID] = editorWidget
		default:
			panic("Not implemented")
		}
	}

	return nil
}

func (e *Executor) StopPage(pageID id.PageID) error {
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
