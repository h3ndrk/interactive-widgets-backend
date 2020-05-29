package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/h3ndrk/containerized-playground/backend/pages"
)

type Handler struct {
	pages    pages.Pages
	mux      *http.ServeMux
	upgrader websocket.Upgrader
}

func NewHandler(runningPages pages.Pages) http.Handler {
	handler := &Handler{
		pages: runningPages,
		mux:   http.NewServeMux(),
	}

	handler.mux.HandleFunc("/pages", func(w http.ResponseWriter, r *http.Request) {
		marshalledPages, err := handler.pages.MarshalPages()
		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(marshalledPages)
	})

	handler.mux.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		pageURL, ok := r.URL.Query()["page_url"]
		if !ok || len(pageURL) < 1 {
			w.WriteHeader(400)
			w.Write([]byte("{\"ok\":false,\"error\":\"URL missing\""))
			return
		}

		marshalledPages, err := handler.pages.MarshalPage(pages.PageURL(pageURL[0]))
		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(marshalledPages)
	})

	handler.mux.HandleFunc("/page/observe", func(w http.ResponseWriter, r *http.Request) {
		pageURL, ok := r.URL.Query()["page_url"]
		if !ok || len(pageURL) < 1 {
			w.WriteHeader(400)
			w.Write([]byte("{\"ok\":false,\"error\":\"URL missing\""))
			return
		}
		roomID, ok := r.URL.Query()["room_id"]
		if !ok || len(roomID) < 1 {
			w.WriteHeader(400)
			w.Write([]byte("{\"ok\":false,\"error\":\"Room ID missing\""))
			return
		}
		pageID, err := pages.PageIDFromPageURLAndRoomID(pages.PageURL(pageURL[0]), pages.RoomID(roomID[0]))
		if err != nil {
			log.Print(err)
			w.WriteHeader(400)
			w.Write([]byte("{\"ok\":false,\"error\":\"Page ID invalid\""))
			return
		}

		connection, err := handler.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		reader := make(chan pages.IncomingMessage)
		writer := make(chan pages.OutgoingMessage)
		observer := pages.ReadWriter{
			Reader: reader,
			Writer: writer,
		}

		err = handler.pages.Observe(pageID, observer)
		if err != nil {
			log.Print(err)
			return
		}

		defer close(reader)

		go func() {
			for message := range writer {
				connection.WriteJSON(&message)
				if err != nil {
					log.Print(err)
					continue
				}
			}
		}()

		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				log.Print(err)
				return
			}

			log.Print(string(message))

			var incomingMessage pages.IncomingMessage
			if err := json.Unmarshal(message, &incomingMessage); err != nil {
				log.Print(err)
				continue
			}

			reader <- incomingMessage
		}
	})

	return handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
