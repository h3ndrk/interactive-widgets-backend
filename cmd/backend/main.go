package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/h3ndrk/inter-md/internal/executor/docker"
	"github.com/h3ndrk/inter-md/internal/multiplexer"
	"github.com/h3ndrk/inter-md/internal/parser"
	"github.com/h3ndrk/inter-md/internal/server"
	"github.com/urfave/cli/v2"
)

func run(c *cli.Context) error {
	pagesDirectoryParser := parser.NewPagesDirectoryParser(c.Path("pages-directory"))
	pages, err := pagesDirectoryParser.GetPages()
	if err != nil {
		return err
	}

	executor, err := docker.NewExecutor(pages)
	if err != nil {
		return err
	}

	multiplexer, err := multiplexer.NewMultiplexer(pages, executor)
	if err != nil {
		return err
	}

	webSocketServer, err := server.NewWebSocketServer(pages, multiplexer)
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:    c.String("listen-address"),
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

	shutdownErrorChannel := make(chan error, 1) // gets closed if goroutine finishes
	go func() {
		defer close(shutdownErrorChannel)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("Got %s", <-signals)

		signal.Stop(dumpSignals)
		close(dumpSignals)

		if err := httpServer.Shutdown(context.Background()); err != nil {
			shutdownErrorChannel <- err
		}
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	webSocketServer.Wait()
	if err := <-shutdownErrorChannel; err != nil {
		return err
	}
	multiplexer.Shutdown()

	return nil
}

func main() {
	pagesDirectoryDefaultValue := "pages"
	currentWorkingDirectory, err := os.Getwd()
	if err == nil {
		pagesDirectoryDefaultValue = filepath.Join(currentWorkingDirectory, "pages")
	}

	app := &cli.App{
		Name:  "backend",
		Usage: "run server backend for inter-md",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "pages-directory",
				Usage:   "absolute path to directory containing pages to serve",
				EnvVars: []string{"PAGES_DIRECTORY"},
				Value:   pagesDirectoryDefaultValue,
			},
			&cli.StringFlag{
				Name:    "listen-address",
				Usage:   "address and port to listen on",
				EnvVars: []string{"LISTEN_ADDRESS"},
				Value:   ":8080",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
