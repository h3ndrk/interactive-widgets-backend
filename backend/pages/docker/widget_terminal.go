package docker

import (
	"encoding/json"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/h3ndrk/containerized-playground/backend/pages/parser"
)

type TerminalWidget struct {
	pageURL     pages.PageURL
	widgetIndex pages.WidgetIndex

	workingDirectory string
}

func NewTerminalWidget(pageURL pages.PageURL, widgetIndex pages.WidgetIndex, widget parser.TerminalWidget) pages.Widget {
	return &TerminalWidget{
		pageURL:          pageURL,
		widgetIndex:      widgetIndex,
		workingDirectory: widget.WorkingDirectory,
	}
}

func (t TerminalWidget) Prepare() error {
	// TODO: build monitor-write image
	return nil
}

func (t TerminalWidget) Cleanup() error {
	return nil
}

func (t TerminalWidget) Instantiate(widgetID pages.WidgetID) (pages.InstantiatedWidget, error) {
	return NewInstantiatedTerminalWidget(widgetID, t.workingDirectory)
}

func (t TerminalWidget) MarshalWidget() ([]byte, error) {
	return json.Marshal(struct {
		Type             string `json:"type"`
		WorkingDirectory string `json:"workingDirectory"`
	}{
		"terminal",
		t.workingDirectory,
	})
}
