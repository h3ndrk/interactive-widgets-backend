package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	webSocketServer, err := server.NewWebSocketServer(pages, multiplexer)
	if err != nil {
		log.Fatal(err)
	}

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: webSocketServer,
	}
	httpServer.RegisterOnShutdown(webSocketServer.Shutdown)

	var shutdownWaiting sync.WaitGroup
	shutdownWaiting.Add(1)
	go func(shutdownWaiting *sync.WaitGroup) {
		defer shutdownWaiting.Done()

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("Got %s", <-signals)

		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}(&shutdownWaiting)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	webSocketServer.Wait()
	shutdownWaiting.Wait()
	multiplexer.Shutdown()
}
