package docker

import (
	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type Page struct {
	widgets []pages.Widget

	pageURL pages.PageURL
}

func (d Page) Prepare() error {
	for widgetID, widget := range d.widgets {
		if err := widget.Prepare(); err != nil {
			return errors.Wrapf(err, "Failed to prepare widget %d", widgetID)
		}
	}
	return nil
}

func (d Page) Cleanup() error {
	for widgetID, widget := range d.widgets {
		if err := widget.Cleanup(); err != nil {
			return errors.Wrapf(err, "Failed to cleanup widget %d", widgetID)
		}
	}
	return nil
}

func (d Page) Instantiate(pageID pages.PageID) (pages.InstantiatedPage, error) {
	return nil, nil
}

func (d Page) MarshalPage() ([]byte, error) {
	return nil, nil
}

func (d Page) MarshalWidgets() ([]byte, error) {
	return nil, nil
}
