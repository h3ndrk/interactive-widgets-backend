package docker

import (
	"encoding/json"

	"github.com/h3ndrk/containerized-playground/backend/pages"
)

type DockerWidgetText struct {
	pageURL     pages.PageURL
	widgetIndex pages.WidgetIndex

	file string
}

func (d DockerWidgetText) Prepare() error {
	// TODO: build monitor-write image
	return nil
}

func (d DockerWidgetText) Cleanup() error {
	return nil
}

func (d DockerWidgetText) Instantiate(widgetID pages.WidgetID) (pages.InstantiatedWidget, error) {
	return NewDockerInstantiatedWidgetText(widgetID, d.file)
}

func (d DockerWidgetText) MarshalWidget() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		File string `json:"file"`
	}{
		"text",
		d.file,
	})
}
