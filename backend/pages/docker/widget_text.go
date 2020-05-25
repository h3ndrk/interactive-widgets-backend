package docker

import (
	"encoding/json"

	"github.com/h3ndrk/containerized-playground/backend/pages"
)

type TextWidget struct {
	pageURL     pages.PageURL
	widgetIndex pages.WidgetIndex

	file string
}

func (d TextWidget) Prepare() error {
	// TODO: build monitor-write image
	return nil
}

func (d TextWidget) Cleanup() error {
	return nil
}

func (d TextWidget) Instantiate(widgetID pages.WidgetID) (pages.InstantiatedWidget, error) {
	return NewInstantiatedTextWidget(widgetID, d.file)
}

func (d TextWidget) MarshalWidget() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		File string `json:"file"`
	}{
		"text",
		d.file,
	})
}
