package docker

import (
	"encoding/json"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/h3ndrk/containerized-playground/backend/pages/parser"
)

type MarkdownWidget struct {
	pageURL     pages.PageURL
	widgetIndex pages.WidgetIndex

	contents string
}

func NewMarkdownWidget(pageURL pages.PageURL, widgetIndex pages.WidgetIndex, widget parser.MarkdownWidget) pages.Widget {
	return &MarkdownWidget{
		pageURL:     pageURL,
		widgetIndex: widgetIndex,
		contents:    widget.Contents,
	}
}

func (m MarkdownWidget) Prepare() error {
	return nil
}

func (m MarkdownWidget) Cleanup() error {
	return nil
}

func (m MarkdownWidget) Instantiate(widgetID pages.WidgetID) (pages.InstantiatedWidget, error) {
	return NewInstantiatedMarkdownWidget(widgetID, m.contents)
}

func (m MarkdownWidget) MarshalWidget() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		File string `json:"contents"`
	}{
		"text",
		m.contents,
	})
}
