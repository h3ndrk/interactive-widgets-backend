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

type attachRequest struct {
	pageID     id.PageID
	client     Client
	errChannel chan error
}

// Multiplexer attaches/detaches multiple clients to one page instance of an
// executor and forwards messages back and forth between the clients and the
// page instance.
type Multiplexer struct {
	executor executor.Executor
	pages    []parser.Page

	attachRequestChannel chan attachRequest
	shutdownWaiting      *sync.WaitGroup
	shutdownChannel      chan struct{}
}

type message struct {
	widgetID id.WidgetID
	data     []byte
}

type detachRequest struct {
	pageID id.PageID
	client Client
}

// NewMultiplexer creates a new multiplexer based on an executor and the
// available pages.
func NewMultiplexer(pages []parser.Page, executor executor.Executor) (*Multiplexer, error) {
	attachRequestChannel := make(chan attachRequest)
	var shutdownWaiting sync.WaitGroup
	shutdownWaiting.Add(1)
	shutdownChannel := make(chan struct{})

	go func(shutdownWaiting *sync.WaitGroup) {
		defer shutdownWaiting.Done()

		startedPages := map[id.PageID][]Client{}
		readFromClientsChannel := make(chan message)
		readFromWidgetsChannel := make(chan message)
		detachRequestChannel := make(chan detachRequest)
		for {
			select {
			case request := <-attachRequestChannel:
				pageURL, roomID, err := id.PageURLAndRoomIDFromPageID(request.pageID)
				if err != nil {
					request.errChannel <- errors.Wrapf(err, "Failed to decode page ID \"%s\"", request.pageID)
					close(request.errChannel)
					break
				}

				page := parser.PageFromPageURL(pages, pageURL)
				if page == nil {
					request.errChannel <- errors.Errorf("No page with URL \"%s\"", pageURL)
					close(request.errChannel)
					break
				}

				if _, ok := startedPages[request.pageID]; !ok {
					if !page.IsInteractive {
						request.errChannel <- errors.Errorf("Page \"%s\" is not interactive", pageURL)
						close(request.errChannel)
						break
					}

					widgetIDs := map[id.WidgetIndex]id.WidgetID{}
					gotError := false
					for widgetIndex, widget := range page.Widgets {
						if !widget.IsInteractive() {
							continue
						}

						widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
						if err != nil {
							request.errChannel <- errors.Wrapf(err, "Failed to encode widget ID for page \"%s\"", request.pageID)
							close(request.errChannel)
							gotError = true
							break
						}

						widgetIDs[id.WidgetIndex(widgetIndex)] = widgetID
					}
					if gotError {
						break
					}

					if err := executor.StartPage(request.pageID); err != nil {
						request.errChannel <- errors.Wrapf(err, "Failed to start page \"%s\"", err)
						close(request.errChannel)
						break
					}

					for widgetIndex, widget := range page.Widgets {
						if !widget.IsInteractive() {
							continue
						}

						widgetID := widgetIDs[id.WidgetIndex(widgetIndex)]

						go func(widget *parser.Widget, widgetID id.WidgetID) {
							for {
								data, err := executor.Read(widgetID)
								if err != nil {
									if err == io.EOF {
										return
									}

									log.Print(err) // there is no error channel to the clients, just log it
									return
								}

								readFromWidgetsChannel <- message{
									widgetID: widgetID,
									data:     data,
								}
							}
						}(&page.Widgets[widgetIndex], widgetID)
					}
				} else {
					// attaching to already started, get current state of all widgets and send it to attaching client
					gotError := false
					for widgetIndex, widget := range page.Widgets {
						if !widget.IsInteractive() {
							continue
						}

						widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(pageURL, roomID, id.WidgetIndex(widgetIndex))
						if err != nil {
							request.errChannel <- errors.Wrapf(err, "Failed to encode widget ID for page \"%s\"", request.pageID)
							close(request.errChannel)
							gotError = true
							break
						}

						data, err := executor.GetCurrentState(widgetID)
						if err != nil {
							request.errChannel <- errors.Wrapf(err, "Failed to get current state of widget \"%s\" for page \"%s\"", widgetID, request.pageID)
							close(request.errChannel)
							gotError = true
							break
						}

						if err := request.client.Write(widgetID, data); err != nil {
							request.errChannel <- errors.Wrapf(err, "Failed to send current state to client for page \"%s\"", request.pageID)
							close(request.errChannel)
							gotError = true
							break
						}
					}
					if gotError {
						break
					}
				}

				startedPages[request.pageID] = append(startedPages[request.pageID], request.client)

				go func() {
					defer func() {
						detachRequestChannel <- detachRequest{
							pageID: request.pageID,
							client: request.client,
						}
					}()

					for {
						widgetID, data, err := request.client.Read()
						if err != nil {
							if err == io.EOF {
								return
							}

							log.Print(err) // there is no error channel to the clients, just log it
							return
						}

						readFromClientsChannel <- message{
							widgetID: widgetID,
							data:     data,
						}
					}
				}()

				close(request.errChannel)
			case message := <-readFromClientsChannel:
				if err := executor.Write(message.widgetID, message.data); err != nil {
					log.Print(err) // there is no error channel to the clients, just log it
				}
			case message := <-readFromWidgetsChannel:
				sendMessageToAttachedClients(message, startedPages)
			case request := <-detachRequestChannel:
				n := 0
				for _, c := range startedPages[request.pageID] {
					if c != request.client {
						startedPages[request.pageID][n] = c
						n++
					}
				}
				startedPages[request.pageID] = startedPages[request.pageID][:n]

				if len(startedPages[request.pageID]) == 0 {
					done := make(chan struct{})
					go func() {
						if err := executor.StopPage(request.pageID); err != nil {
							log.Print(err) // there is no error channel to the clients, just log it
							// continue to remove it
						}
						done <- struct{}{}
					}()

					// while waiting for StopPage, further process readFromWidgetsChannel (and drain the stopping page)
				loop:
					for {
						select {
						case message := <-readFromWidgetsChannel:
							sendMessageToAttachedClients(message, startedPages)
						case <-done:
							break loop
						}
					}

					delete(startedPages, request.pageID)
				}
			case <-shutdownChannel:
				if len(startedPages) > 0 {
					log.Print("Bug: Multiplexer termination request while pages are running.")
				}

				return
			}
		}
	}(&shutdownWaiting)

	return &Multiplexer{
		executor:             executor,
		pages:                pages,
		attachRequestChannel: attachRequestChannel,
		shutdownWaiting:      &shutdownWaiting,
		shutdownChannel:      shutdownChannel,
	}, nil
}

func sendMessageToAttachedClients(message message, startedPages map[id.PageID][]Client) {
	pageURL, roomID, _, err := id.PageURLAndRoomIDAndWidgetIndexFromWidgetID(message.widgetID)
	if err != nil {
		log.Print(err) // there is no error channel to the clients, just log it
		return
	}
	pageID, err := id.PageIDFromPageURLAndRoomID(pageURL, roomID)
	if err != nil {
		log.Print(err) // there is no error channel to the clients, just log it
		return
	}

	var writeWaiting sync.WaitGroup
	writeWaiting.Add(len(startedPages[pageID]))
	for clientIndex := range startedPages[pageID] {
		go func(clientIndex int, writeWaiting *sync.WaitGroup) {
			defer writeWaiting.Done()

			if err := startedPages[pageID][clientIndex].Write(message.widgetID, message.data); err != nil {
				log.Print(err) // there is no error channel to the clients, just log it
			}
		}(clientIndex, &writeWaiting)
	}

	writeWaiting.Wait()
}

// Attach attaches a given client to a given page ID. If the page instance
// already exists, it is attached to it. Otherwise a new page instance is
// started. When the client closes, the client is detached from the page
// instance. If there are no more attached clients left, the page instance gets
// stopped. It is safe to call this function with same page ID arguments from
// different goroutines.
func (m *Multiplexer) Attach(pageID id.PageID, client Client) error {
	errChannel := make(chan error)
	m.attachRequestChannel <- attachRequest{
		pageID:     pageID,
		client:     client,
		errChannel: errChannel,
	}

	if err := <-errChannel; err != nil {
		return errors.Wrapf(err, "Failed to attach client to page \"%s\"", pageID)
	}

	return nil
}

// Shutdown terminates the multiplexer and waits for termination.
func (m *Multiplexer) Shutdown() {
	close(m.shutdownChannel)
	m.shutdownWaiting.Wait()
}
