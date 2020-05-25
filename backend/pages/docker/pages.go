package docker

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type ObservedInstantiatedPage struct {
	instantiatedPage     pages.InstantiatedPage
	addObserverWriter    chan chan<- pages.Message
	removeObserverWriter chan chan<- pages.Message
}

type Pages struct {
	pages                          map[pages.PageURL]pages.Page
	observedInstantiatedPagesMutex sync.Mutex
	observedInstantiatedPages      map[pages.PageID]ObservedInstantiatedPage
	observedInstantiatedPagesLocks PageLock
}

func NewPages() pages.Pages {
	return &Pages{
		pages: map[pages.PageURL]pages.Page{
			"/": Page{
				widgets: []pages.Widget{
					TextWidget{
						pageURL:     "/",
						widgetIndex: 0,
						file:        "/data/test.txt",
					}, TextWidget{
						pageURL:     "/",
						widgetIndex: 1,
						file:        "/data/test.txt",
					}, TextWidget{
						pageURL:     "/",
						widgetIndex: 2,
						file:        "/data/test.txt",
					}, TextWidget{
						pageURL:     "/",
						widgetIndex: 3,
						file:        "/data/test.txt",
					}, TextWidget{
						pageURL:     "/",
						widgetIndex: 4,
						file:        "/data/test.txt",
					},
				},
				pageURL: "/",
			},
			"/run": Page{
				widgets: []pages.Widget{
					TextWidget{
						pageURL:     "/run",
						widgetIndex: 0,
						file:        "/data/test.txt",
					}, TextWidget{
						pageURL:     "/run",
						widgetIndex: 1,
						file:        "/data/test.txt",
					}, TextWidget{
						pageURL:     "/run",
						widgetIndex: 2,
						file:        "/data/test.txt",
					}, TextWidget{
						pageURL:     "/run",
						widgetIndex: 3,
						file:        "/data/test.txt",
					}, TextWidget{
						pageURL:     "/run",
						widgetIndex: 4,
						file:        "/data/test.txt",
					},
				},
				pageURL: "/run",
			},
		},
		observedInstantiatedPages: map[pages.PageID]ObservedInstantiatedPage{},
		observedInstantiatedPagesLocks: PageLock{
			pageLocks: map[pages.PageID]*sync.Mutex{},
		},
	}
}

func (p Pages) Prepare() error {
	for pageURL, page := range p.pages {
		if err := page.Prepare(); err != nil {
			return errors.Wrapf(err, "Failed to prepare page %s", pageURL)
		}
	}

	return nil
}

func (p Pages) Cleanup() error {
	for pageURL, page := range p.pages {
		if err := page.Cleanup(); err != nil {
			return errors.Wrapf(err, "Failed to cleanup page %s", pageURL)
		}
	}

	return nil
}

func (p *Pages) Observe(pageID pages.PageID, observer pages.ReadWriter) error {
	pageURL, _, err := pages.PageURLAndRoomIDFromPageID(pageID)
	if err != nil {
		return nil
	}
	page, ok := p.pages[pageURL]
	if !ok {
		return errors.Errorf("Page URL \"%s\" not existing", pageURL)
	}

	// first, blockingly lock this page (this prevents concurrent instantiation/teardown of pages; once teardown is in progress, the lock is kept as long as teardown is in progress)
	p.observedInstantiatedPagesLocks.Lock(pageID)
	defer p.observedInstantiatedPagesLocks.Unlock(pageID)

	// second, lock the whole map of instantiated pages
	p.observedInstantiatedPagesMutex.Lock()
	defer p.observedInstantiatedPagesMutex.Unlock()

	// at this point: no teardown of this page and no other modification of the map of instantiated pages is in progress
	// we can now modify the map of instantiated pages and can safely add observers to the current page (will be added to an instantiated page, not to a page that is currently in teardown)
	instantiatedPage, ok := p.observedInstantiatedPages[pageID]
	if !ok {
		newInstantiatedPage, err := page.Instantiate(pageID)
		if err != nil {
			return nil
		}

		addObserverWriter := make(chan chan<- pages.Message)
		removeObserverWriter := make(chan chan<- pages.Message)
		instantiatedPage = ObservedInstantiatedPage{
			instantiatedPage:     newInstantiatedPage,
			addObserverWriter:    addObserverWriter,
			removeObserverWriter: removeObserverWriter,
		}
		p.observedInstantiatedPages[pageID] = instantiatedPage

		// the following goroutine manages all messages and channels that address sending from an instantiated page to all connected observers
		go func() {
			var writers []chan<- pages.Message
			for {
				select {
				case message, ok := <-instantiatedPage.instantiatedPage.GetReader():
					if !ok {
						defer p.observedInstantiatedPagesLocks.Unlock(pageID)

						// page closed channel: remove page, disconnect all observer writers
						// this lock will not introduce any deadlock, since we currently have the lock of this page
						p.observedInstantiatedPagesMutex.Lock()
						defer p.observedInstantiatedPagesMutex.Unlock()

						// the reader of this page should only close if there are no observer writers left
						if len(writers) > 0 {
							panic("Bug: An instantiated page which is currently in teardown cannot have any active observer writers.")
						}

						// remove page
						delete(p.observedInstantiatedPages, pageID)
						return
					}

					for _, writer := range writers {
						writer <- message
					}
				case writer := <-addObserverWriter:
					writers = append(writers, writer)
				case writer := <-removeObserverWriter:
					// remove writer
					n := 0
					for _, w := range writers {
						if w == writer {
							writers[n] = w
							n++
						}
					}
					writers = writers[:n]

					// close writer
					close(writer)

					// if last writer closed, close page (keep page lock locked, will be unlocked once the page reader closes), else unlock page lock
					if len(writers) == 0 {
						close(instantiatedPage.instantiatedPage.GetWriter())
					} else {
						p.observedInstantiatedPagesLocks.Unlock(pageID)
					}
				}
			}
		}()
	}

	instantiatedPage.addObserverWriter <- observer.Writer

	go func() {
		// read from observer
		for message := range observer.Reader {
			instantiatedPage.instantiatedPage.GetWriter() <- message
		}

		// observer reader closed, start locking this page and also remove this observer writer
		p.observedInstantiatedPagesLocks.Lock(pageID)
		instantiatedPage.removeObserverWriter <- observer.Writer
	}()

	return nil
}

func (p Pages) MarshalPages() ([]byte, error) {
	var pages [][]byte
	for pageURL, page := range p.pages {
		page, err := page.MarshalPage()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to marshal page %s", pageURL)
		}

		pages = append(pages, page)
	}

	return []byte(fmt.Sprintf("{\"pages\":[%s]}", bytes.Join(pages, []byte(",")))), nil
}

func (p Pages) MarshalPage(pageURL pages.PageURL) ([]byte, error) {
	page, ok := p.pages[pageURL]
	if !ok {
		return nil, errors.Errorf("Page URL \"%s\" not existing", pageURL)
	}

	return page.MarshalWidgets()
}
