package multiplexer

import (
	"io"
	"log"
	"sync"

	"github.com/h3ndrk/containerized-playground/internal/executor"
	"github.com/h3ndrk/containerized-playground/internal/id"
	"github.com/h3ndrk/containerized-playground/internal/parser"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Multiplexer attaches/detaches multiple clients to one page instance of an
// executor and forwards messages back and forth between the clients and the
// page instance.
type Multiplexer struct {
	executor executor.Executor
	pages    []parser.Page

	startedPagesMutex sync.Mutex
	startedPagesLocks pageLock
	startedPages      map[id.PageID][]Client
}

type message struct {
	WidgetID id.WidgetID
	Data     []byte
}

// NewMultiplexer creates a new multiplexer based on an executor and the
// available pages.
func NewMultiplexer(pages []parser.Page, executor executor.Executor) (*Multiplexer, error) {
	// TODO: add context to all New*() functions to be able to tear down the whole process
	return &Multiplexer{
		executor: executor,
		pages:    pages,
		startedPagesLocks: pageLock{
			pageLocks: map[id.PageID]*sync.Mutex{},
		},
		startedPages: map[id.PageID][]Client{},
	}, nil
}

// Attach attaches a given client to a given page ID. If the page instance
// already exists, it is attached to it. Otherwise a new page instance is
// started. When the client closes, the client is detached from the page
// instance. If there are no more attached clients left, the page instance gets
// stopped. It is safe to call this function with same page ID arguments from
// different goroutines.
func (m *Multiplexer) Attach(pageID id.PageID, client Client) error {
	pageURL, roomID, err := id.PageURLAndRoomIDFromPageID(pageID)
	if err != nil {
		return errors.Wrapf(err, "Failed to decode page ID \"%s\"", pageID)
	}

	page := parser.PageFromPageURL(m.pages, pageURL)
	if page == nil {
		return errors.Errorf("No page with URL \"%s\"", pageURL)
	}

	if !page.IsInteractive {
		return errors.Errorf("Page \"%s\" is not interactive", pageURL)
	}

	// First, blockingly lock this page (this prevents concurrent instantiation
	// or teardown of pages; once teardown is in progress, the lock is kept as
	// long as teardown is in progress)
	m.startedPagesLocks.lock(pageID)
	defer m.startedPagesLocks.unlock(pageID)

	// Second, lock the whole map of started pages
	m.startedPagesMutex.Lock()
	defer m.startedPagesMutex.Unlock()

	// At this point: No teardown of this page and no other modification of the
	// map of started pages is in progress. We can now modify the map of
	// started pages and can safely add clients to the current page (will be
	// added to an started page, not to a page that is currently in teardown).
	attachedClients, ok := m.startedPages[pageID]
	if !ok {
		if err := m.executor.StartPage(pageID); err != nil {
			return errors.Wrapf(err, "Failed to start page \"%s\"", err)
		}

		// stores the zero value (no attached clients) in the started pages
		m.startedPages[pageID] = attachedClients

		// connect all widgets of the page to mux channel
		widgetIDs := map[id.WidgetIndex]id.WidgetID{}
		for widgetIndex, widget := range page.Widgets {
			if !widget.IsInteractive() {
				continue
			}

			var cumulatedErrors error
			widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
			if err != nil {
				cumulatedErrors = multierror.Append(cumulatedErrors, errors.Wrapf(err, "Failed to encode widget ID for page \"%s\"", pageID))

				if err := m.executor.StopPage(pageID); err != nil {
					cumulatedErrors = multierror.Append(cumulatedErrors, errors.Wrapf(err, "Failed to stop page \"%s\"", pageID))
					// continue to remove it
				}

				// remove page
				delete(m.startedPages, pageID)

				return cumulatedErrors
			}

			widgetIDs[id.WidgetIndex(widgetIndex)] = widgetID
		}

		for widgetIndex, widget := range page.Widgets {
			if !widget.IsInteractive() {
				continue
			}

			widgetID := widgetIDs[id.WidgetIndex(widgetIndex)]

			go func(widget *parser.Widget, widgetID id.WidgetID) {
				for {
					data, err := m.executor.Read(widgetID)
					if err != nil {
						if err == io.EOF {
							return
						}

						log.Print(err) // there is no error channel to the clients, just log it
						return
					}

					// lock page to allow safe access of attached clients
					m.startedPagesLocks.lock(pageID)

					var writeWaiting sync.WaitGroup
					writeWaiting.Add(len(attachedClients))
					for clientIndex := range attachedClients {
						go func(clientIndex int, writeWaiting *sync.WaitGroup) {
							defer writeWaiting.Done()

							if err := attachedClients[clientIndex].Write(widgetID, data); err != nil {
								log.Print(err) // there is no error channel to the clients, just log it
							}
						}(clientIndex, &writeWaiting)
					}

					m.startedPagesLocks.unlock(pageID)

					writeWaiting.Wait()
				}
			}(&page.Widgets[widgetIndex], widgetID)
		}
	}

	attachedClients = append(attachedClients, client)

	go func() {
		defer func() {
			m.startedPagesLocks.lock(pageID)
			defer m.startedPagesLocks.unlock(pageID)

			// detach client
			n := 0
			for _, c := range attachedClients {
				if c != client {
					attachedClients[n] = c
					n++
				}
			}
			attachedClients = attachedClients[:n]

			// if last client closed, close page
			if len(attachedClients) == 0 {
				// this lock will not introduce any deadlock, since we currently have the lock of this page (correct locking order)
				m.startedPagesMutex.Lock()
				defer m.startedPagesMutex.Unlock()

				if err := m.executor.StopPage(pageID); err != nil {
					log.Print(err) // there is no error channel to the clients, just log it
					// continue to remove it
				}

				// remove page
				delete(m.startedPages, pageID)
			}
		}()

		for {
			widgetID, data, err := client.Read()
			if err != nil {
				if err == io.EOF {
					return
				}

				log.Print(err) // there is no error channel to the clients, just log it
				return
			}

			if err := m.executor.Write(widgetID, data); err != nil {
				log.Print(err) // there is no error channel to the clients, just log it
				continue
			}
		}
	}()

	return nil
}
