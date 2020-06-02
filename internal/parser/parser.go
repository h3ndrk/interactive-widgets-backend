package parser

import "github.com/h3ndrk/containerized-playground/internal/id"

type Parser interface {
	GetPages() ([]Page, error)
}

// PageFromPageURL searches given pages for a page with given page URL.
// Returns a pointer to the found page or nil.
func PageFromPageURL(pages []Page, pageURL id.PageURL) *Page {
	for i, page := range pages {
		if page.URL == pageURL {
			return &pages[i]
		}
	}

	return nil
}

type PageMetadata struct {
	IsInteractive bool       `json:"isInteractive"`
	BasePath      string     `json:"-"`
	URL           id.PageURL `json:"url"`
}

type Page struct {
	PageMetadata
	Widgets    []Widget `json:"widgets"`
	ImagePaths []string `json:"imagePaths"`
}

type Widget interface {
	IsInteractive() bool
}

type MarkdownWidget struct {
	Contents string `json:"contents"`
}

func (MarkdownWidget) IsInteractive() bool {
	return false
}

type TextWidget struct {
	File string `json:"file"`
}

func (TextWidget) IsInteractive() bool {
	return true
}

type ImageWidget struct {
	File string `json:"file"`
}

func (ImageWidget) IsInteractive() bool {
	return true
}

type ButtonWidget struct {
	Label   string `json:"label"`
	Command string `json:"command"`
}

func (ButtonWidget) IsInteractive() bool {
	return true
}

type EditorWidget struct {
	File string `json:"file"`
}

func (EditorWidget) IsInteractive() bool {
	return true
}

type TerminalWidget struct {
	WorkingDirectory string `json:"workingDirectory"`
}

func (TerminalWidget) IsInteractive() bool {
	return true
}
