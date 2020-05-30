package main

import (
	"log"
	"net/http"

	"github.com/h3ndrk/containerized-playground/internal/executor/docker"
	"github.com/h3ndrk/containerized-playground/internal/multiplexer"
	"github.com/h3ndrk/containerized-playground/internal/parser"
	"github.com/h3ndrk/containerized-playground/internal/server"
)

func main() {
	pagesDirectoryParser := parser.NewPagesDirectoryParser("pages")
	pages, err := pagesDirectoryParser.GetPages()
	if err != nil {
		log.Fatal(err)
	}

	executor, err := docker.NewExecutor(pages)
	if err != nil {
		log.Fatal(err)
	}

	multiplexer, err := multiplexer.NewMultiplexer(pages, executor)
	if err != nil {
		log.Fatal(err)
	}

	server, err := server.NewWebSocketServer(pages, multiplexer)
	if err != nil {
		log.Fatal(err)
	}

	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatal(err)
	}
}
