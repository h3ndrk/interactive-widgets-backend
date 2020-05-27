package docker

import (
	"encoding/json"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/h3ndrk/containerized-playground/backend/pages/parser"
)

type TextWidget struct {
	pageURL     pages.PageURL
	widgetIndex pages.WidgetIndex

	file string
}

func NewTextWidget(pageURL pages.PageURL, widgetIndex pages.WidgetIndex, widget parser.TextWidget) pages.Widget {
	return &TextWidget{
		pageURL:     pageURL,
		widgetIndex: widgetIndex,
		file:        widget.File,
	}
}

func (t TextWidget) Prepare() error {
	// TODO: build monitor-write image
	return nil
}

func (t TextWidget) Cleanup() error {
	return nil
}

func (t TextWidget) Instantiate(widgetID pages.WidgetID) (pages.InstantiatedWidget, error) {
	return NewInstantiatedTextWidget(widgetID, t.file)
}

func (t TextWidget) MarshalWidget() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		File string `json:"file"`
	}{
		"text",
		t.file,
	})
}
