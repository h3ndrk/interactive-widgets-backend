package docker

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type Page struct {
	widgets []pages.Widget

	pageURL pages.PageURL
}

func (p Page) Prepare() error {
	for widgetIndex, widget := range p.widgets {
		if err := widget.Prepare(); err != nil {
			return errors.Wrapf(err, "Failed to prepare widget %d", widgetIndex)
		}
	}

	return nil
}

func (p Page) Cleanup() error {
	for widgetIndex, widget := range p.widgets {
		if err := widget.Cleanup(); err != nil {
			return errors.Wrapf(err, "Failed to cleanup widget %d", widgetIndex)
		}
	}

	return nil
}

func (p Page) Instantiate(pageID pages.PageID) (pages.InstantiatedPage, error) {
	return nil, nil
}

func (p Page) MarshalPage() ([]byte, error) {
	return json.Marshal(p.pageURL)
}

func (p Page) MarshalWidgets() ([]byte, error) {
	pageURL, err := json.Marshal(p.pageURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal page URL")
	}

	var widgets [][]byte
	for widgetIndex, widget := range p.widgets {
		widget, err := widget.MarshalWidget()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to marshal widget %d", widgetIndex)
		}

		widgets = append(widgets, widget)
	}

	return []byte(fmt.Sprintf("{\"pageUrl\":%s,\"widgets\":[%s]}", pageURL, bytes.Join(widgets, []byte(",")))), nil
}
