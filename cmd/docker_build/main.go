package main

import (
	"log"

	"github.com/h3ndrk/inter-md/internal/executor/docker"
	"github.com/h3ndrk/inter-md/internal/parser"
)

func main() {
	pagesDirectoryParser := parser.NewPagesDirectoryParser("pages")
	pages, err := pagesDirectoryParser.GetPages()
	if err != nil {
		log.Fatal(err)
	}

	executor, err := docker.NewExecutor(pages, []string{}, []string{}, []string{})
	if err != nil {
		log.Fatal(err)
	}

	dockerExecutor, ok := executor.(*docker.Executor)
	if !ok {
		log.Fatal("Failed to assert type of docker executor")
	}

	if err := dockerExecutor.BuildImages(); err != nil {
		log.Fatal(err)
	}
}
