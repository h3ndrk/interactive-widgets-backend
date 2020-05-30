package multiplexer

import (
	"io"
	"log"
	"sync"

	"github.com/h3ndrk/containerized-playground/internal/executor"
	"github.com/h3ndrk/containerized-playground/internal/id"
	"github.com/h3ndrk/containerized-playground/internal/parser"
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
	startedPages      map[id.PageID]startedPage
}

type message struct {
	WidgetID id.WidgetID
	Data     []byte
}

type startedPage struct {
	attachedClients []Client
	// toWidget is a mux channel from all attached clients to the widgets of the page (n:1)
	toWidget chan message
	// toClients is a mux channel from all widgets of the page to clients (n:1)
	toClients chan message
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
		startedPages: map[id.PageID]startedPage{},
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
		return err
	}

	page := parser.PageFromPageURL(m.pages, pageURL)
	if page == nil {
		return errors.Errorf("No page with URL %s", pageURL)
	}

	if !page.IsInteractive {
		return errors.Errorf("Page %s is not interactive", pageURL)
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
	pageInstance, ok := m.startedPages[pageID]
	if !ok {
		if err := m.executor.StartPage(pageID); err != nil {
			return nil
		}

		pageInstance = startedPage{
			toWidget:  make(chan message),
			toClients: make(chan message),
		}
		m.startedPages[pageID] = pageInstance

		// connect all widgets of the page to mux channel
		var fromPageCloseWaiting sync.WaitGroup
		for widgetIndex, widget := range page.Widgets {
			if !widget.IsInteractive() {
				continue
			}

			widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
			if err != nil {
				// TODO: handle error (remove partially created widget readers), if we reach this code, this is a bug. the error should be handled in executor.StartPage
				return err
			}

			fromPageCloseWaiting.Add(1)
			go func(widget *parser.Widget, widgetID id.WidgetID, fromPageCloseWaiting *sync.WaitGroup) {
				defer fromPageCloseWaiting.Done()

				for {
					data, err := m.executor.Read(widgetID)
					if err != nil {
						if err == io.EOF {
							return
						}

						log.Print(err)
						return
					}

					pageInstance.toClients <- message{
						WidgetID: widgetID,
						Data:     data,
					}
				}
			}(&page.Widgets[widgetIndex], widgetID, &fromPageCloseWaiting)
		}

		// this goroutine closes the mux channel once all widgets closed
		go func() {
			fromPageCloseWaiting.Wait()
			close(pageInstance.toClients)
		}()

		// this goroutine sends all messages from widgets to all attached clients
		go func() {
			for message := range pageInstance.toClients {
				// lock page to allow safe access of attached clients
				m.startedPagesLocks.lock(pageID)
				defer m.startedPagesLocks.unlock(pageID)

				var writeWaiting sync.WaitGroup
				writeWaiting.Add(len(pageInstance.attachedClients))
				for clientIndex := range pageInstance.attachedClients {
					go func(clientIndex int, writeWaiting *sync.WaitGroup) {
						defer writeWaiting.Done()

						if err := pageInstance.attachedClients[clientIndex].Write(message.WidgetID, message.Data); err != nil {
							log.Print(err)
							// TODO: better error handling
						}
					}(clientIndex, &writeWaiting)
				}

				writeWaiting.Wait()
			}
		}()

		// this goroutine sends messages to the widgets in the executor
		go func() {
			for message := range pageInstance.toWidget {
				if err := m.executor.Write(message.WidgetID, message.Data); err != nil {
					log.Print(err)
					continue
				}
			}

			// at this point the channel closed while having the lock of the page, therefore ensure it is released at the end
			defer m.startedPagesLocks.unlock(pageID)

			// this lock will not introduce any deadlock, since we currently have the lock of this page (correct locking order)
			m.startedPagesMutex.Lock()
			defer m.startedPagesMutex.Unlock()

			// the reader of this page should only close if there are no observer writers left
			if len(pageInstance.attachedClients) > 0 {
				panic("Bug: An instantiated page which is currently in teardown cannot have any attached client.")
			}

			if err := m.executor.StopPage(pageID); err != nil {
				log.Print(err)
				// continue to remove it
				// TODO: make error handling better
			}

			// remove page
			delete(m.startedPages, pageID)
		}()
	}

	pageInstance.attachedClients = append(pageInstance.attachedClients, client)

	go func() {
		defer func() {
			m.startedPagesLocks.lock(pageID)

			// detach client
			n := 0
			for _, c := range pageInstance.attachedClients {
				if c != client {
					pageInstance.attachedClients[n] = c
					n++
				}
			}
			pageInstance.attachedClients = pageInstance.attachedClients[:n]

			// if last client closed, close page (keep page lock locked, will be unlocked once the page reader closes), else unlock page lock
			if len(pageInstance.attachedClients) == 0 {
				close(pageInstance.toWidget)
			} else {
				m.startedPagesLocks.unlock(pageID)
			}
		}()

		for {
			widgetID, data, err := client.Read()
			if err != nil {
				if err == io.EOF {
					return
				}

				log.Print(err)
				return
			}

			pageInstance.toWidget <- message{
				WidgetID: widgetID,
				Data:     data,
			}
		}
	}()

	return nil
}
