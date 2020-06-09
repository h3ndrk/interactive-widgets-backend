package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gabriel-vasile/mimetype"
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

	shutdownWaiting *sync.WaitGroup
	shutdownChannel chan struct{}
}

// NewWebSocketServer creates a new websocket server which serves the given
// pages metadata/widgets and allows attaching to the given multiplexer.
func NewWebSocketServer(pages []parser.Page, multiplexer *multiplexer.Multiplexer) (Server, error) {
	server := &WebSocketServer{
		pages:           pages,
		multiplexer:     multiplexer,
		mux:             http.NewServeMux(),
		shutdownWaiting: &sync.WaitGroup{},
		shutdownChannel: make(chan struct{}),
	}

	// serve list of pages with their metadata
	server.mux.HandleFunc("/pages", func(w http.ResponseWriter, r *http.Request) {
		var pageMetadatas []parser.PageMetadata
		for _, page := range server.pages {
			pageMetadatas = append(pageMetadatas, page.PageMetadata)
		}

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(pageMetadatas); err != nil {
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
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(page); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})

	// serve attachment endpoint which upgrades to a websocket connection
	server.mux.HandleFunc("/page/attach", func(w http.ResponseWriter, r *http.Request) {
		server.shutdownWaiting.Add(1)
		defer server.shutdownWaiting.Done()

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
			writeMutex:    &sync.Mutex{},
		}
		client.closeWaiting.Add(1)

		clientCloseChannel := make(chan struct{})
		go func() {
			select {
			case <-server.shutdownChannel:
				client.writeMutex.Lock()
				defer client.writeMutex.Unlock()
				connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "shutdown"))
				time.Sleep(time.Second)
				connection.Close()
			case <-clientCloseChannel:
			}
		}()

		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := connection.WriteMessage(websocket.PingMessage, nil); err != nil {
						log.Print(err)
						return
					}
				case <-clientCloseChannel:
					return
				}
			}
		}()

		if err := server.multiplexer.Attach(pageID, &client); err != nil {
			log.Print(err) // there is no error channel to the client, just log it
		}

		client.closeWaiting.Wait()
		close(clientCloseChannel)
	})

	// serve images embedded in markdown of pages
	server.mux.HandleFunc("/page/image", func(w http.ResponseWriter, r *http.Request) {
		pageURLQueryParam, ok := r.URL.Query()["page_url"]
		if !ok || len(pageURLQueryParam) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Query param \"page_url\" missing"))
			return
		}
		pageURL := id.PageURL(pageURLQueryParam[0])

		imagePathQueryParam, ok := r.URL.Query()["image_path"]
		if !ok || len(imagePathQueryParam) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Query param \"image_path\" missing"))
			return
		}
		imagePath := imagePathQueryParam[0]

		page := parser.PageFromPageURL(server.pages, pageURL)
		if page == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("No page with URL \"%s\"", pageURL)))
			return
		}

		foundImagePath := false
		for _, imagePathOfPage := range page.ImagePaths {
			if imagePath == imagePathOfPage {
				foundImagePath = true
				break
			}
		}
		if !foundImagePath {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("No image with path \"%s\" found in page with URL \"%s\"", imagePath, pageURL)))
			return
		}

		imagePathBasedOnPageDirectory := filepath.Join(page.BasePath, imagePath)

		allowedMIMETypes := []string{"image/gif", "image/vnd.microsoft.icon", "image/jpeg", "image/png", "image/svg+xml", "image/tiff", "image/webp"}
		mime, err := mimetype.DetectFile(imagePathBasedOnPageDirectory)
		if err != nil {
			log.Print(err)
			// mime remains valid
		}
		if !mimetype.EqualsAny(mime.String(), allowedMIMETypes...) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(fmt.Sprintf("Image path \"%s\" is not an image (MIME: %s)", imagePath, mime.String())))
			return
		}

		image, err := os.Open(imagePathBasedOnPageDirectory)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		defer image.Close()
		w.Header().Set("Content-Type", mime.String())
		io.Copy(w, image)
	})

	return server, nil
}

func (s *WebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Shutdown terminates the websocket server.
func (s *WebSocketServer) Shutdown() {
	close(s.shutdownChannel)
}

// Wait waits for the websocket server to terminate.
func (s *WebSocketServer) Wait() {
	s.shutdownWaiting.Wait()
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
	writeMutex    *sync.Mutex
}

func (c *webSocketClient) Read() (id.WidgetID, []byte, error) {
	_, message, err := c.connection.ReadMessage()
	if err != nil {
		c.closeWaiting.Done()

		if _, ok := err.(*websocket.CloseError); ok {
			return "", nil, io.EOF
		}

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

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	if err := c.connection.WriteJSON(&webSocketMessage{
		WidgetIndex: widgetIndex,
		Data:        data,
	}); err != nil {
		return err
	}

	return nil
}
