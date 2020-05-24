package docker

import (
	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type DockerPage struct {
	widgets []pages.Widget

	pageURL pages.PageURL
}

func (d DockerPage) Prepare() error {
	for widgetID, widget := range d.widgets {
		if err := widget.Prepare(); err != nil {
			return errors.Wrapf(err, "Failed to prepare widget %d", widgetID)
		}
	}
	return nil
}

func (d DockerPage) Cleanup() error {
	for widgetID, widget := range d.widgets {
		if err := widget.Cleanup(); err != nil {
			return errors.Wrapf(err, "Failed to cleanup widget %d", widgetID)
		}
	}
	return nil
}

func (d DockerPage) Instantiate(pageID pages.PageID) (pages.InstantiatedPage, error) {
	return nil, nil
}

func (d DockerPage) MarshalPage() ([]byte, error) {
	return nil, nil
}

func (d DockerPage) MarshalWidgets() ([]byte, error) {
	return nil, nil
}
