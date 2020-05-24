package docker

import (
	"sync"

	"github.com/h3ndrk/containerized-playground/backend/pages"
	"github.com/pkg/errors"
)

type ObservedInstantiatedPage struct {
	instantiatedPage     pages.InstantiatedPage
	addObserverWriter    chan chan<- pages.Message
	removeObserverWriter chan chan<- pages.Message
}

type DockerPages struct {
	pages                          map[pages.PageURL]pages.Page
	observedInstantiatedPagesMutex sync.Mutex
	observedInstantiatedPages      map[pages.PageID]ObservedInstantiatedPage
	observedInstantiatedPagesLocks PageLock
}

func NewDockerPages() pages.Pages {
	return &DockerPages{
		pages: map[pages.PageURL]pages.Page{
			"/": DockerPage{
				widgets: []pages.Widget{
					DockerWidget{
						widgetIndex: 0,
					}, DockerWidget{
						widgetIndex: 1,
					}, DockerWidget{
						widgetIndex: 2,
					}, DockerWidget{
						widgetIndex: 3,
					}, DockerWidget{
						widgetIndex: 4,
					},
				},
				pageURL: "/",
			},
			"/run": DockerPage{
				widgets: []pages.Widget{
					DockerWidget{
						widgetIndex: 0,
					}, DockerWidget{
						widgetIndex: 1,
					}, DockerWidget{
						widgetIndex: 2,
					}, DockerWidget{
						widgetIndex: 3,
					}, DockerWidget{
						widgetIndex: 4,
					},
				},
				pageURL: "/run",
			},
		},
		observedInstantiatedPages: map[pages.PageID]ObservedInstantiatedPage{},
		observedInstantiatedPagesLocks: PageLock{
			pageLocks: map[pages.PageID]sync.Mutex{},
		},
	}
}

func (d DockerPages) Prepare() error {
	for pageID, page := range d.pages {
		if err := page.Prepare(); err != nil {
			return errors.Wrapf(err, "Failed to prepare page %s", pageID)
		}
	}
	return nil
}

func (d DockerPages) Cleanup() error {
	for pageID, page := range d.pages {
		if err := page.Cleanup(); err != nil {
			return errors.Wrapf(err, "Failed to cleanup page %s", pageID)
		}
	}
	return nil
}

func (d *DockerPages) Observe(pageID pages.PageID, observer pages.ReadWriter) error {
	pageURL, _, err := pages.PageURLAndRoomIDFromPageID(pageID)
	if err != nil {
		return nil
	}
	page, ok := d.pages[pageURL]
	if !ok {
		return errors.Errorf("pages.Page URL \"%s\" not existing", pageURL)
	}
	// first, blockingly lock this page (this prevents concurrent instantiation/teardown of pages; once teardown is in progress, the lock is kept as long as teardown is in progress)
	d.observedInstantiatedPagesLocks.Lock(pageID)
	defer d.observedInstantiatedPagesLocks.Unlock(pageID)
	// second, lock the whole map of instantiated pages
	d.observedInstantiatedPagesMutex.Lock()
	defer d.observedInstantiatedPagesMutex.Unlock()
	// at this point: no teardown of this page and no other modification of the map of instantiated pages is in progress
	// we can now modify the map of instantiated pages and can safely add observers to the current page (will be added to an instantiated page, not to a page that is currently in teardown)
	instantiatedPage, ok := d.observedInstantiatedPages[pageID]
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
		d.observedInstantiatedPages[pageID] = instantiatedPage
		// the following goroutine manages all messages and channels that address sending from an instantiated page to all connected observers
		go func() {
			var writers []chan<- pages.Message
			for {
				select {
				case message, ok := <-instantiatedPage.instantiatedPage.GetReader():
					if !ok {
						defer d.observedInstantiatedPagesLocks.Unlock(pageID)
						// page closed channel: remove page, disconnect all observer writers
						// this lock will not introduce any deadlock, since we currently have the lock of this page
						d.observedInstantiatedPagesMutex.Lock()
						defer d.observedInstantiatedPagesMutex.Unlock()
						// the reader of this page should only close if there are no observer writers left
						if len(writers) > 0 {
							panic("Bug: An instantiated page which is currently in teardown cannot have any active observer writers.")
						}
						// remove page
						delete(d.observedInstantiatedPages, pageID)
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
						d.observedInstantiatedPagesLocks.Unlock(pageID)
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
		d.observedInstantiatedPagesLocks.Lock(pageID)
		instantiatedPage.removeObserverWriter <- observer.Writer
	}()
	return nil
}

func (d DockerPages) MarshalPages() ([]byte, error) {
	// TODO
	return nil, nil
}

func (d DockerPages) MarshalPage(pageURL pages.PageURL) ([]byte, error) {
	// TODO
	return nil, nil
}
