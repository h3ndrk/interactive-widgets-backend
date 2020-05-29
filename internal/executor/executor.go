package executor

import "github.com/h3ndrk/containerized-playground/internal/id"

type Executor interface {
	StartPage(id.PageID) error
	StopPage(id.PageID) error
	Read(id.WidgetID) ([]byte, error)
	Write(id.WidgetID, []byte) error
}
