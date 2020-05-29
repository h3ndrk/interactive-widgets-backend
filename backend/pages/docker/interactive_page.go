package docker

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/h3ndrk/containerized-playground/backend/pages/parser"
	"github.com/pkg/errors"
)

type InteractivePage struct {
	widgets []pages.Widget

	pageURL pages.PageURL
}

func NewInteractivePage(pageURL pages.PageURL, readPage parser.Page) (pages.Page, error) {
	if !readPage.IsInteractive {
		return nil, errors.New("Got non-interactive read page")
	}
	if len(readPage.Widgets) == 0 {
		return nil, errors.New("Not containing any widgets")
	}

	var widgets []pages.Widget
	for widgetIndex, readWidget := range readPage.Widgets {
		switch readWidget := readWidget.(type) {
		case parser.MarkdownWidget:
			widgets = append(widgets, NewMarkdownWidget(pageURL, pages.WidgetIndex(widgetIndex), readWidget))
		case parser.TextWidget:
			widgets = append(widgets, NewTextWidget(pageURL, pages.WidgetIndex(widgetIndex), readWidget))
		case parser.ImageWidget:
			widgets = append(widgets, NewImageWidget(pageURL, pages.WidgetIndex(widgetIndex), readWidget))
		case parser.ButtonWidget:
			widgets = append(widgets, NewButtonWidget(pageURL, pages.WidgetIndex(widgetIndex), readWidget))
		case parser.EditorWidget:
			widgets = append(widgets, NewEditorWidget(pageURL, pages.WidgetIndex(widgetIndex), readWidget))
		case parser.TerminalWidget:
			widgets = append(widgets, NewTerminalWidget(pageURL, pages.WidgetIndex(widgetIndex), readWidget))
		default:
			return nil, errors.Errorf("Got unimplemented widget type %T", readWidget)
		}
	}

	return &InteractivePage{
		widgets: widgets,
		pageURL: pageURL,
	}, nil
}

func (i InteractivePage) Prepare() error {
	for widgetIndex, widget := range i.widgets {
		if err := widget.Prepare(); err != nil {
			return errors.Wrapf(err, "Failed to prepare widget %d", widgetIndex)
		}
	}

	return nil
}

func (i InteractivePage) Cleanup() error {
	for widgetIndex, widget := range i.widgets {
		if err := widget.Cleanup(); err != nil {
			return errors.Wrapf(err, "Failed to cleanup widget %d", widgetIndex)
		}
	}

	return nil
}

func (i InteractivePage) Instantiate(pageID pages.PageID) (pages.InstantiatedPage, error) {
	return NewInstantiatedPage(pageID, i.widgets)
}

func (i InteractivePage) MarshalPage() ([]byte, error) {
	return json.Marshal(i.pageURL)
}

func (i InteractivePage) MarshalWidgets() ([]byte, error) {
	pageURL, err := json.Marshal(i.pageURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal page URL")
	}

	var widgets [][]byte
	for widgetIndex, widget := range i.widgets {
		widget, err := widget.MarshalWidget()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to marshal widget %d", widgetIndex)
		}

		widgets = append(widgets, widget)
	}

	return []byte(fmt.Sprintf("{\"isInteractive\":true,\"pageUrl\":%s,\"widgets\":[%s]}", pageURL, bytes.Join(widgets, []byte(",")))), nil
}
