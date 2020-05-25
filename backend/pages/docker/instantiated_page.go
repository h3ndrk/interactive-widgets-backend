package docker

import "github.com/h3ndrk/containerized-playground/backend/pages"

type InstantiatedPage struct {
	instantiatedWidgets []pages.InstantiatedWidget

	pageID pages.PageID
}
