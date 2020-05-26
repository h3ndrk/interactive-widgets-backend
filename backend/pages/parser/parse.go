package parser

import (
	"bytes"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type WidgetType string

const MarkdownWidgetType WidgetType = "markdown"
const TextWidgetType WidgetType = "text"
const ImageWidgetType WidgetType = "image"
const ButtonWidgetType WidgetType = "button"
const EditorWidgetType WidgetType = "editor"
const TerminalWidgetType WidgetType = "terminal"

type Widget interface {
	GetWidgetType() WidgetType
	getBegin() int
	getEnd() int
}

type MarkdownWidget struct {
	Contents string
	begin    int
	end      int
}

func (MarkdownWidget) GetWidgetType() WidgetType {
	return MarkdownWidgetType
}

func (m MarkdownWidget) getBegin() int {
	return m.begin
}

func (m MarkdownWidget) getEnd() int {
	return m.end
}

type TextWidget struct {
	File  string
	begin int
	end   int
}

func (TextWidget) GetWidgetType() WidgetType {
	return TextWidgetType
}

func (t TextWidget) getBegin() int {
	return t.begin
}

func (t TextWidget) getEnd() int {
	return t.end
}

type ImageWidget struct {
	File  string
	begin int
	end   int
}

func (ImageWidget) GetWidgetType() WidgetType {
	return ImageWidgetType
}

func (i ImageWidget) getBegin() int {
	return i.begin
}

func (i ImageWidget) getEnd() int {
	return i.end
}

type ButtonWidget struct {
	Label   string
	Command string
	begin   int
	end     int
}

func (ButtonWidget) GetWidgetType() WidgetType {
	return ButtonWidgetType
}

func (b ButtonWidget) getBegin() int {
	return b.begin
}

func (b ButtonWidget) getEnd() int {
	return b.end
}

type EditorWidget struct {
	File  string
	begin int
	end   int
}

func (EditorWidget) GetWidgetType() WidgetType {
	return EditorWidgetType
}

func (e EditorWidget) getBegin() int {
	return e.begin
}

func (e EditorWidget) getEnd() int {
	return e.end
}

type TerminalWidget struct {
	WorkingDirectory string
	begin            int
	end              int
}

func (TerminalWidget) GetWidgetType() WidgetType {
	return TerminalWidgetType
}

func (t TerminalWidget) getBegin() int {
	return t.begin
}

func (t TerminalWidget) getEnd() int {
	return t.end
}

func processBlock(contents []byte, block ast.Node) (Widget, error) {
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
		return nil, errors.Wrapf(err, "Failed to parse HTML: %s", blockContent)
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
			return nil, errors.Wrapf(err, "Missing \"file\" attribute in text widget: %s", blockContent)
		}

		return TextWidget{
			File:  attributes["file"],
			begin: blockStart,
			end:   blockStop,
		}, nil
	case "image":
		if _, ok := attributes["file"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"file\" attribute in image widget: %s", blockContent)
		}

		return ImageWidget{
			File:  attributes["file"],
			begin: blockStart,
			end:   blockStop,
		}, nil
	case "button":
		if _, ok := attributes["command"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"command\" attribute in button widget: %s", blockContent)
		}
		if interactiveElement.FirstChild == nil || interactiveElement.FirstChild.Type != html.TextNode {
			return nil, errors.Wrapf(err, "Missing label in button widget: %s", blockContent)
		}

		return ButtonWidget{
			Label:   interactiveElement.FirstChild.Data,
			Command: attributes["command"],
			begin:   blockStart,
			end:     blockStop,
		}, nil
	case "editor":
		if _, ok := attributes["file"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"file\" attribute in editor widget: %s", blockContent)
		}

		return EditorWidget{
			File:  attributes["file"],
			begin: blockStart,
			end:   blockStop,
		}, nil
	case "terminal":
		if _, ok := attributes["working-directory"]; !ok {
			return nil, errors.Wrapf(err, "Missing \"working-directory\" attribute in terminal widget: %s", blockContent)
		}

		return TerminalWidget{
			WorkingDirectory: attributes["working-directory"],
			begin:            blockStart,
			end:              blockStop,
		}, nil
	default:
		return nil, errors.Wrapf(err, "Unknown widget: %s", blockContent)
	}
}

func fillGaps(contents []byte, widgets []Widget) []Widget {
	var widgetsWithoutGaps []Widget

	// eventually add markdown widget to fill gap
	trimmedGap := bytes.TrimSpace(contents[0:widgets[0].getBegin()])
	if len(trimmedGap) > 0 {
		widgetsWithoutGaps = append(widgetsWithoutGaps, MarkdownWidget{
			Contents: string(trimmedGap),
		})
	}
	for widgetIndex, widget := range widgets {
		// add current widget
		widgetsWithoutGaps = append(widgetsWithoutGaps, widget)

		// eventually add markdown widget to fill gap
		gapEnd := len(contents)
		if widgetIndex+1 < len(widgets) {
			gapEnd = widgets[widgetIndex+1].getBegin()
		}
		trimmedGap := bytes.TrimSpace(contents[widget.getEnd():gapEnd])
		if len(trimmedGap) > 0 {
			widgetsWithoutGaps = append(widgetsWithoutGaps, MarkdownWidget{
				Contents: string(trimmedGap),
			})
		}
	}

	return widgetsWithoutGaps
}

func ParsePage(pagePath string) ([]Widget, error) {
	contents, err := ioutil.ReadFile(pagePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read page %s", pagePath)
	}

	md := goldmark.New()
	reader := text.NewReader(contents)
	document := md.Parser().Parse(reader)

	var widgets []Widget
	for block := document.FirstChild(); block != nil; block = block.NextSibling() {
		widget, err := processBlock(contents, block)
		if err != nil {
			return nil, err
		}
		if widget != nil {
			widgets = append(widgets, widget)
		}
	}

	if len(widgets) == 0 {
		return []Widget{
			MarkdownWidget{
				Contents: string(contents),
			},
		}, nil
	}

	return fillGaps(contents, widgets), nil
}
