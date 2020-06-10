package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"

	"github.com/h3ndrk/inter-md/internal/executor/docker"
	"github.com/h3ndrk/inter-md/internal/multiplexer"
	"github.com/h3ndrk/inter-md/internal/parser"
	"github.com/h3ndrk/inter-md/internal/server"
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

	dumpSignals := make(chan os.Signal, 1)
	signal.Notify(dumpSignals, syscall.SIGUSR1)
	go func() {
		for range dumpSignals {
			fmt.Fprintf(os.Stderr, "Current #goroutines: %d\n\n", runtime.NumGoroutine())
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 2)
		}
	}()

	var shutdownWaiting sync.WaitGroup
	shutdownWaiting.Add(1)
	go func(shutdownWaiting *sync.WaitGroup) {
		defer shutdownWaiting.Done()

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("Got %s", <-signals)

		signal.Stop(dumpSignals)
		close(dumpSignals)

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
