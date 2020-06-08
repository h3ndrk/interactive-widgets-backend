package parser

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/h3ndrk/containerized-playground/internal/id"
	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type PageDirectoryParser struct {
	pagesDirectory string
}

func NewPagesDirectoryParser(pagesDirectory string) Parser {
	return &PageDirectoryParser{
		pagesDirectory: pagesDirectory,
	}
}

func (p *PageDirectoryParser) GetPages() ([]Page, error) {
	var readPages []Page

	err := filepath.Walk(p.pagesDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "Failed to access path \"%s\"", path)
		}

		if !info.IsDir() && info.Name() == "page.md" {
			basePath := filepath.Dir(path)
			relativeBasePath, err := filepath.Rel(p.pagesDirectory, basePath)
			if err != nil {
				return errors.Wrapf(err, "Failed to create relative base path of page \"%s\"", path)
			}
			url := filepath.Join(string(filepath.Separator), relativeBasePath)

			dockerfilePath := filepath.Join(basePath, "Dockerfile")
			dockerfileExists := true
			_, err = os.Stat(dockerfilePath)
			if err != nil {
				dockerfileExists = false
				if !os.IsNotExist(err) {
					return errors.Wrapf(err, "Failed to access path \"%s\"", dockerfilePath)
				}
			}

			title, widgets, imagePaths, err := parsePage(path)
			if err != nil {
				return errors.Wrapf(err, "Failed to parse page \"%s\"", path)
			}

			for _, imagePath := range imagePaths {
				if !strings.HasPrefix(filepath.Join(basePath, imagePath), p.pagesDirectory) {
					return errors.Errorf("Image path \"%s\" of page \"%s\" escapes pages directory \"%s\"", imagePath, url, p.pagesDirectory)
				}
			}

			readPages = append(readPages, Page{
				PageMetadata: PageMetadata{
					IsInteractive: dockerfileExists,
					BasePath:      basePath,
					URL:           id.PageURL(url),
					Title:         title,
				},
				Widgets:    widgets,
				ImagePaths: imagePaths,
			})
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read pages in \"%s\"", p.pagesDirectory)
	}

	return readPages, nil
}

func parsePage(pagePath string) (string, []Widget, []string, error) {
	contents, err := ioutil.ReadFile(pagePath)
	if err != nil {
		return "", nil, nil, errors.Wrapf(err, "Failed to read page \"%s\"", pagePath)
	}

	md := goldmark.New()
	reader := text.NewReader(contents)
	document := md.Parser().Parse(reader)

	var title string
	var titleEnd int
	var widgetsWithSlice []widgetWithSlice
	var imagePaths []string
	for block := document.FirstChild(); block != nil; block = block.NextSibling() {
		if block == document.FirstChild() {
			if block.Kind() != ast.KindHeading {
				return "", nil, nil, errors.New("First markdown block is not a heading")
			}
			if block.(*ast.Heading).Level != 1 {
				return "", nil, nil, errors.New("First markdown block is not a level 1 heading")
			}

			lines := block.Lines()
			if lines.Len() == 0 {
				return "", nil, nil, errors.New("First markdown block has no lines")
			}

			title = string(block.Text(contents))
			titleEnd = lines.At(lines.Len() - 1).Stop

			continue
		}

		widget, err := processBlock(contents, block)
		if err != nil {
			return "", nil, nil, err
		}
		if widget != nil {
			widgetsWithSlice = append(widgetsWithSlice, *widget)
		}

		imagePaths = append(imagePaths, extractImagePaths(block)...)
	}

	if len(widgetsWithSlice) == 0 {
		return title, []Widget{
			MarkdownWidget{
				Type:     "markdown",
				Contents: string(bytes.TrimSpace(contents[titleEnd:])),
			},
		}, imagePaths, nil
	}

	return title, fillGaps(contents, widgetsWithSlice, titleEnd), imagePaths, nil
}

type widgetWithSlice struct {
	widget Widget
	begin  int
	end    int
}

func processBlock(contents []byte, block ast.Node) (*widgetWithSlice, error) {
	lines := block.Lines()
	if lines.Len() == 0 {
		// ignoring blocks with no reference to source contents (e.g. text blocks, list blocks)
		return nil, nil
	}

	// filter invalid HTML blocks (only allowed: Markdown paragraph with a single HTML element)
	blockStart := lines.At(0).Start
	blockStop := lines.At(lines.Len() - 1).Stop
	blockContent := contents[blockStart:blockStop]
	parsedBlock, err := html.Parse(bytes.NewReader(blockContent))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse HTML: \"%s\"", blockContent)
	}
	if parsedBlock.Type != html.DocumentNode {
		return nil, nil
	}
	if parsedBlock.FirstChild == nil || parsedBlock.FirstChild.Type != html.ElementNode || parsedBlock.FirstChild.Data != "html" || parsedBlock.FirstChild.DataAtom != atom.Html {
		return nil, nil
	}
	htmlElement := parsedBlock.FirstChild
	if htmlElement.FirstChild == nil || htmlElement.FirstChild.Type != html.ElementNode || htmlElement.FirstChild.Data != "head" || htmlElement.FirstChild.DataAtom != atom.Head {
		return nil, nil
	}
	headElement := htmlElement.FirstChild
	if headElement.NextSibling == nil || headElement.NextSibling.Type != html.ElementNode || headElement.NextSibling.Data != "body" || headElement.NextSibling.DataAtom != atom.Body {
		return nil, nil
	}
	bodyElement := headElement.NextSibling
	if bodyElement.FirstChild == nil || bodyElement.FirstChild.Type != html.ElementNode || bodyElement.FirstChild.Data[0:2] != "x-" {
		return nil, nil
	}
	interactiveElement := bodyElement.FirstChild
	if interactiveElement.NextSibling != nil {
		return nil, nil
	}

	attributes := map[string]string{}
	for _, attribute := range interactiveElement.Attr {
		// only store the first non-existing attribute key
		if _, ok := attributes[attribute.Key]; !ok {
			attributes[attribute.Key] = attribute.Val
		}
	}

	switch interactiveElement.Data[2:] {
	case "text":
		if _, ok := attributes["file"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"file\" attribute in text widget: \"%s\"", blockContent)
		}

		return &widgetWithSlice{
			widget: &TextWidget{
				Type: interactiveElement.Data[2:],
				File: attributes["file"],
			},
			begin: blockStart,
			end:   blockStop,
		}, nil
	case "image":
		if _, ok := attributes["file"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"file\" attribute in image widget: \"%s\"", blockContent)
		}
		if _, ok := attributes["mime"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"mime\" attribute in image widget: \"%s\"", blockContent)
		}

		return &widgetWithSlice{
			widget: &ImageWidget{
				Type: interactiveElement.Data[2:],
				File: attributes["file"],
				MIME: attributes["mime"],
			},
			begin: blockStart,
			end:   blockStop,
		}, nil
	case "button":
		if _, ok := attributes["command"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"command\" attribute in button widget: \"%s\"", blockContent)
		}
		if interactiveElement.FirstChild == nil || interactiveElement.FirstChild.Type != html.TextNode {
			return nil, errors.Wrapf(err, "Missing label in button widget: \"%s\"", blockContent)
		}

		return &widgetWithSlice{
			widget: &ButtonWidget{
				Type:    interactiveElement.Data[2:],
				Label:   interactiveElement.FirstChild.Data,
				Command: attributes["command"],
			},
			begin: blockStart,
			end:   blockStop,
		}, nil
	case "editor":
		if _, ok := attributes["file"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"file\" attribute in editor widget: \"%s\"", blockContent)
		}

		return &widgetWithSlice{
			widget: &EditorWidget{
				Type: interactiveElement.Data[2:],
				File: attributes["file"],
			},
			begin: blockStart,
			end:   blockStop,
		}, nil
	case "terminal":
		if _, ok := attributes["working-directory"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"working-directory\" attribute in terminal widget: \"%s\"", blockContent)
		}

		return &widgetWithSlice{
			widget: &TerminalWidget{
				Type:             interactiveElement.Data[2:],
				WorkingDirectory: attributes["working-directory"],
			},
			begin: blockStart,
			end:   blockStop,
		}, nil
	default:
		return nil, errors.Wrapf(err, "Unknown widget: \"%s\"", blockContent)
	}
}

func fillGaps(contents []byte, widgets []widgetWithSlice, offset int) []Widget {
	var widgetsWithoutGaps []Widget

	// eventually add markdown widget to fill gap
	trimmedGap := bytes.TrimSpace(contents[offset:widgets[0].begin])
	if len(trimmedGap) > 0 {
		widgetsWithoutGaps = append(widgetsWithoutGaps, MarkdownWidget{
			Type:     "markdown",
			Contents: string(trimmedGap),
		})
	}
	for widgetIndex, widget := range widgets {
		// add current widget
		widgetsWithoutGaps = append(widgetsWithoutGaps, widget.widget)

		// eventually add markdown widget to fill gap
		gapEnd := len(contents)
		if widgetIndex+1 < len(widgets) {
			gapEnd = widgets[widgetIndex+1].begin
		}
		trimmedGap := bytes.TrimSpace(contents[widget.end:gapEnd])
		if len(trimmedGap) > 0 {
			widgetsWithoutGaps = append(widgetsWithoutGaps, MarkdownWidget{
				Type:     "markdown",
				Contents: string(trimmedGap),
			})
		}
	}

	return widgetsWithoutGaps
}

func extractImagePaths(block ast.Node) []string {
	var imagePaths []string

	if block.Kind() == ast.KindImage {
		imagePaths = append(imagePaths, string(block.(*ast.Image).Destination))
	}

	for child := block.FirstChild(); child != nil; child = child.NextSibling() {
		imagePaths = append(imagePaths, extractImagePaths(child)...)
	}

	return imagePaths
}
