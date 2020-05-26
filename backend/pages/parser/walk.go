package parser

import (
	"os"
	"path/filepath"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type Page struct {
	IsInteractive bool
	BasePath      string
	URL           pages.PageURL
	Widgets       []Widget
	ImagePaths    []string
}

func ReadPages(pagesDirectory string) ([]Page, error) {
	var readPages []Page

	err := filepath.Walk(pagesDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "Failed to access path %s", path)
		}

		if !info.IsDir() && info.Name() == "page.md" {
			basePath := filepath.Dir(path)
			relativeBasePath, err := filepath.Rel(pagesDirectory, basePath)
			if err != nil {
				return errors.Wrapf(err, "Failed to create relative base path of page %s", path)
			}
			url := filepath.Join(string(filepath.Separator), relativeBasePath)

			dockerfilePath := filepath.Join(basePath, "Dockerfile")
			dockerfileExists := true
			_, err = os.Stat(dockerfilePath)
			if err != nil {
				dockerfileExists = false
				if !os.IsNotExist(err) {
					return errors.Wrapf(err, "Failed to access path %s", dockerfilePath)
				}
			}

			widgets, imagePaths, err := ParsePage(path)
			if err != nil {
				return errors.Wrapf(err, "Failed to parse page %s", path)
			}

			readPages = append(readPages, Page{
				IsInteractive: dockerfileExists,
				BasePath:      basePath,
				URL:           pages.PageURL(url),
				Widgets:       widgets,
				ImagePaths:    imagePaths,
			})
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read pages in %s", pagesDirectory)
	}

	return readPages, nil
}
