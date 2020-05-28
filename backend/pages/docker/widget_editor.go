package docker

import (
	"encoding/json"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/h3ndrk/containerized-playground/backend/pages/parser"
)

type EditorWidget struct {
	pageURL     pages.PageURL
	widgetIndex pages.WidgetIndex

	file string
}

func NewEditorWidget(pageURL pages.PageURL, widgetIndex pages.WidgetIndex, widget parser.EditorWidget) pages.Widget {
	return &EditorWidget{
		pageURL:     pageURL,
		widgetIndex: widgetIndex,
		file:        widget.File,
	}
}

func (e EditorWidget) Prepare() error {
	// TODO: build monitor-write image
	return nil
}

func (e EditorWidget) Cleanup() error {
	return nil
}

func (e EditorWidget) Instantiate(widgetID pages.WidgetID) (pages.InstantiatedWidget, error) {
	return NewInstantiatedEditorWidget(widgetID, e.file)
}

func (e EditorWidget) MarshalWidget() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		File string `json:"file"`
	}{
		"editor",
		e.file,
	})
}