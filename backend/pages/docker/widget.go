package docker

import "github.com/h3ndrk/containerized-playground/backend/pages"

type DockerWidget struct {
	pageURL     pages.PageURL
	widgetIndex pages.WidgetIndex
}

func (d DockerWidget) Prepare() error {
	return nil
}

func (d DockerWidget) Cleanup() error {
	return nil
}

func (d DockerWidget) Instantiate(widgetID pages.WidgetID) (pages.InstantiatedWidget, error) {
	return nil, nil
}

func (d DockerWidget) MarshalWidget() ([]byte, error) {
	return nil, nil
}
