package docker

import "github.com/h3ndrk/containerized-playground/backend/pages"

type DockerInstantiatedPage struct {
	instantiatedWidgets []pages.InstantiatedWidget

	pageID pages.PageID
}
