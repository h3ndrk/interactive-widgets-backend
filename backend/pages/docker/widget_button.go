package docker

import (
	"encoding/json"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/h3ndrk/containerized-playground/backend/pages/parser"
)

type ButtonWidget struct {
	pageURL     pages.PageURL
	widgetIndex pages.WidgetIndex

	label   string
	command string
}

func NewButtonWidget(pageURL pages.PageURL, widgetIndex pages.WidgetIndex, widget parser.ButtonWidget) pages.Widget {
	return &ButtonWidget{
		pageURL:     pageURL,
		widgetIndex: widgetIndex,
		label:       widget.Label,
		command:     widget.Command,
	}
}

func (b ButtonWidget) Prepare() error {
	// TODO: build monitor-write image
	return nil
}

func (b ButtonWidget) Cleanup() error {
	return nil
}

func (b ButtonWidget) Instantiate(widgetID pages.WidgetID) (pages.InstantiatedWidget, error) {
	return NewInstantiatedButtonWidget(widgetID, b.command)
}

func (b ButtonWidget) MarshalWidget() ([]byte, error) {
	return json.Marshal(struct {
		Type    string `json:"type"`
		Label   string `json:"label"`
		Command string `json:"command"`
	}{
		"Button",
		b.label,
		b.command,
	})
}
