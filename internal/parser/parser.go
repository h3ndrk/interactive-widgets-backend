package parser

import "github.com/h3ndrk/inter-md/internal/id"

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
	Title         string     `json:"title"`
}

type Page struct {
	PageMetadata
	Widgets    []Widget `json:"widgets"`
	ImagePaths []string `json:"-"`
}

type Widget interface {
	IsInteractive() bool
}

type MarkdownWidget struct {
	Type     string `json:"type"`
	Contents string `json:"contents"`
}

func (MarkdownWidget) IsInteractive() bool {
	return false
}

type TextWidget struct {
	Type string `json:"type"`
	File string `json:"file"`
}

func (TextWidget) IsInteractive() bool {
	return true
}

type ImageWidget struct {
	Type string `json:"type"`
	File string `json:"file"`
	MIME string `json:"mime"`
}

func (ImageWidget) IsInteractive() bool {
	return true
}

type ButtonWidget struct {
	Type    string `json:"type"`
	Label   string `json:"label"`
	Command string `json:"command"`
}

func (ButtonWidget) IsInteractive() bool {
	return true
}

type EditorWidget struct {
	Type string `json:"type"`
	File string `json:"file"`
}

func (EditorWidget) IsInteractive() bool {
	return true
}

type TerminalWidget struct {
	Type             string `json:"type"`
	WorkingDirectory string `json:"workingDirectory"`
}

func (TerminalWidget) IsInteractive() bool {
	return true
}
