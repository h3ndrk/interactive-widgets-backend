package docker

import (
	"encoding/json"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/h3ndrk/containerized-playground/backend/pages/parser"
)

type ImageWidget struct {
	pageURL     pages.PageURL
	widgetIndex pages.WidgetIndex

	file string
}

func NewImageWidget(pageURL pages.PageURL, widgetIndex pages.WidgetIndex, widget parser.ImageWidget) pages.Widget {
	return &ImageWidget{
		pageURL:     pageURL,
		widgetIndex: widgetIndex,
		file:        widget.File,
	}
}

func (i ImageWidget) Prepare() error {
	// TODO: build monitor-write image
	return nil
}

func (i ImageWidget) Cleanup() error {
	return nil
}

func (i ImageWidget) Instantiate(widgetID pages.WidgetID) (pages.InstantiatedWidget, error) {
	return NewInstantiatedImageWidget(widgetID, i.file)
}

func (i ImageWidget) MarshalWidget() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		File string `json:"file"`
	}{
		"Image",
		i.file,
	})
}
