package docker

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type InteractivePage struct {
	widgets []pages.Widget

	pageURL pages.PageURL
}

func (i InteractivePage) Prepare() error {
	for widgetIndex, widget := range i.widgets {
		if err := widget.Prepare(); err != nil {
			return errors.Wrapf(err, "Failed to prepare widget %d", widgetIndex)
		}
	}

	return nil
}

func (i InteractivePage) Cleanup() error {
	for widgetIndex, widget := range i.widgets {
		if err := widget.Cleanup(); err != nil {
			return errors.Wrapf(err, "Failed to cleanup widget %d", widgetIndex)
		}
	}

	return nil
}

func (i InteractivePage) Instantiate(pageID pages.PageID) (pages.InstantiatedPage, error) {
	return NewInstantiatedPage(pageID, i.widgets)
}

func (i InteractivePage) MarshalPage() ([]byte, error) {
	return json.Marshal(i.pageURL)
}

func (i InteractivePage) MarshalWidgets() ([]byte, error) {
	pageURL, err := json.Marshal(i.pageURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal page URL")
	}

	var widgets [][]byte
	for widgetIndex, widget := range i.widgets {
		widget, err := widget.MarshalWidget()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to marshal widget %d", widgetIndex)
		}

		widgets = append(widgets, widget)
	}

	return []byte(fmt.Sprintf("{\"pageUrl\":%s,\"widgets\":[%s]}", pageURL, bytes.Join(widgets, []byte(",")))), nil
}
