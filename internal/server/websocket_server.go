package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/h3ndrk/containerized-playground/internal/id"
	"github.com/h3ndrk/containerized-playground/internal/multiplexer"
	"github.com/h3ndrk/containerized-playground/internal/parser"
	"github.com/pkg/errors"
)

// WebSocketServer serves a websocket interface over HTTP that attaches clients
// to a multiplexer.
type WebSocketServer struct {
	pages       []parser.Page
	multiplexer *multiplexer.Multiplexer

	mux      *http.ServeMux
	upgrader websocket.Upgrader
}

// NewWebSocketServer creates a new websocket server which serves the given
// pages metadata/widgets and allows attaching to the given multiplexer.
func NewWebSocketServer(pages []parser.Page, multiplexer *multiplexer.Multiplexer) (Server, error) {
	server := &WebSocketServer{
		pages:       pages,
		multiplexer: multiplexer,
		mux:         http.NewServeMux(),
	}

	// serve list of pages with their metadata
	server.mux.HandleFunc("/pages", func(w http.ResponseWriter, r *http.Request) {
		var pageMetadatas []parser.PageMetadata
		for _, page := range server.pages {
			pageMetadatas = append(pageMetadatas, page.PageMetadata)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(pageMetadatas); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})

	// serve page with it's widgets
	server.mux.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		pageURLQueryParam, ok := r.URL.Query()["page_url"]
		if !ok || len(pageURLQueryParam) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Query param \"page_url\" missing"))
			return
		}
		pageURL := id.PageURL(pageURLQueryParam[0])

		page := parser.PageFromPageURL(server.pages, pageURL)
		if page == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("No page with URL \"%s\"", pageURL)))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(page); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})

	// serve attachment endpoint which upgrades to a websocket connection
	server.mux.HandleFunc("/page/attach", func(w http.ResponseWriter, r *http.Request) {
		pageURLQueryParam, ok := r.URL.Query()["page_url"]
		if !ok || len(pageURLQueryParam) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Query param \"page_url\" missing"))
			return
		}
		pageURL := id.PageURL(pageURLQueryParam[0])

		roomIDQueryParam, ok := r.URL.Query()["room_id"]
		if !ok || len(roomIDQueryParam) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Query param \"room_id\" missing"))
			return
		}
		roomID := id.RoomID(roomIDQueryParam[0])

		pageID, err := id.PageIDFromPageURLAndRoomID(pageURL, roomID)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Given \"page_url\" and \"room_id\" are invalid"))
			return
		}

		page := parser.PageFromPageURL(server.pages, pageURL)
		if page == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("No page with URL \"%s\"", pageURL)))
			return
		}

		if !page.IsInteractive {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(fmt.Sprintf("Cannot attach to non-interactive page with URL \"%s\"", pageURL)))
			return
		}

		connection, err := server.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print(err) // there is no error channel to the client, just log it
			return
		}

		client := webSocketClient{
			numberOfPages: len(page.Widgets),
			pageURL:       pageURL,
			roomID:        roomID,
			pageID:        pageID,
			connection:    connection,
			closeWaiting:  &sync.WaitGroup{},
		}
		client.closeWaiting.Add(1)

		if err := server.multiplexer.Attach(pageID, &client); err != nil {
			log.Print(err) // there is no error channel to the client, just log it
		}

		client.closeWaiting.Wait()
	})

	return server, nil
}

func (s *WebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type webSocketMessage struct {
	WidgetIndex id.WidgetIndex  `json:"widgetIndex"`
	Data        json.RawMessage `json:"data"`
}

type webSocketClient struct {
	numberOfPages int
	pageURL       id.PageURL
	roomID        id.RoomID
	pageID        id.PageID
	connection    *websocket.Conn
	closeWaiting  *sync.WaitGroup
}

func (c *webSocketClient) Read() (id.WidgetID, []byte, error) {
	_, message, err := c.connection.ReadMessage()
	if err != nil {
		c.closeWaiting.Done()
		return "", nil, err
	}

	var unmarshalledMessage webSocketMessage
	if err := json.Unmarshal(message, &unmarshalledMessage); err != nil {
		c.closeWaiting.Done()
		return "", nil, err
	}

	if unmarshalledMessage.WidgetIndex < 0 || int(unmarshalledMessage.WidgetIndex) >= c.numberOfPages {
		c.closeWaiting.Done()
		return "", nil, errors.Errorf("Widget index %d at page \"%s\" is out of range", unmarshalledMessage.WidgetIndex, c.pageID)
	}

	widgetID, err := id.WidgetIDFromPageURLAndRoomIDAndWidgetIndex(c.pageURL, c.roomID, unmarshalledMessage.WidgetIndex)
	if err != nil {
		c.closeWaiting.Done()
		return "", nil, err
	}

	return widgetID, unmarshalledMessage.Data, nil
}

func (c *webSocketClient) Write(widgetID id.WidgetID, data []byte) error {
	pageURL, roomID, widgetIndex, err := id.PageURLAndRoomIDAndWidgetIndexFromWidgetID(widgetID)
	if err != nil {
		return err
	}

	if pageURL != c.pageURL || roomID != c.roomID {
		return errors.Errorf("Widget ID \"%s\" of message to send to client is invalid", widgetID)
	}

	if err := c.connection.WriteJSON(&webSocketMessage{
		WidgetIndex: widgetIndex,
		Data:        data,
	}); err != nil {
		return err
	}

	return nil
}
