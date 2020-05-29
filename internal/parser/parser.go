package parser

import "github.com/h3ndrk/containerized-playground/internal/id"

type Parser interface {
	GetPages() ([]Page, error)
}

type Page struct {
	IsInteractive bool       `json:"isInteractive"`
	BasePath      string     `json:"basePath"`
	URL           id.PageURL `json:"url"`
	Widgets       []Widget   `json:"widgets"`
	ImagePaths    []string   `json:"imagePaths"`
}

type Widget interface {
	IsWidget()
}

type MarkdownWidget struct {
	Contents string `json:"contents"`
}

func (MarkdownWidget) IsWidget() {}

type TextWidget struct {
	File string `json:"file"`
}

func (TextWidget) IsWidget() {}

type ImageWidget struct {
	File string `json:"file"`
}

func (ImageWidget) IsWidget() {}

type ButtonWidget struct {
	Label   string `json:"label"`
	Command string `json:"command"`
}

func (ButtonWidget) IsWidget() {}

type EditorWidget struct {
	File string `json:"file"`
}

func (EditorWidget) IsWidget() {}

type TerminalWidget struct {
	WorkingDirectory string `json:"workingDirectory"`
}

func (TerminalWidget) IsWidget() {}
