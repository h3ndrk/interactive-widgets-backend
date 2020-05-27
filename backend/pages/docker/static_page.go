package docker

import (
	"encoding/json"
	"fmt"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/h3ndrk/containerized-playground/backend/pages/parser"
	"github.com/pkg/errors"
)

type StaticPage struct {
	markdownWidget MarkdownWidget

	pageURL pages.PageURL
}

func NewStaticPage(pageURL pages.PageURL, readPage parser.Page) (pages.Page, error) {
	if readPage.IsInteractive {
		return nil, errors.New("Got interactive read page")
	}
	if len(readPage.Widgets) != 1 {
		return nil, errors.New("Not containing one widget")
	}
	if readPage.Widgets[0].GetWidgetType() != parser.MarkdownWidgetType {
		return nil, errors.New("Wrong widget type: Not a markdown widget")
	}
	readMarkdownWidget, ok := readPage.Widgets[0].(parser.MarkdownWidget)
	if !ok {
		return nil, errors.New("Wrong widget type: Not a markdown widget (type assertion)")
	}
	markdownWidget, ok := NewMarkdownWidget(pageURL, 0, readMarkdownWidget).(MarkdownWidget)
	if !ok {
		return nil, errors.New("Wrong widget type constructed: Not a markdown widget")
	}

	return &StaticPage{
		markdownWidget: markdownWidget,
		pageURL:        pageURL,
	}, nil
}

func (s StaticPage) Prepare() error {
	return nil
}

func (s StaticPage) Cleanup() error {
	return nil
}

func (s StaticPage) Instantiate(pageID pages.PageID) (pages.InstantiatedPage, error) {
	panic("Bug: Cannot instantiate static page.")
}

func (s StaticPage) MarshalPage() ([]byte, error) {
	return json.Marshal(s.pageURL)
}

func (s StaticPage) MarshalWidgets() ([]byte, error) {
	pageURL, err := json.Marshal(s.pageURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal page URL")
	}

	marshalledWidget, err := s.markdownWidget.MarshalWidget()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal widget 0")
	}

	return []byte(fmt.Sprintf("{\"isInteractive\":false,\"pageUrl\":%s,\"widgets\":[%s]}", pageURL, marshalledWidget)), nil
}
