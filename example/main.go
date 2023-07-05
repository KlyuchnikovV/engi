package main

import (
	"log"

	"github.com/KlyuchnikovV/webapi"
	"github.com/KlyuchnikovV/webapi/example/services"
	"github.com/KlyuchnikovV/webapi/types"
)

func main() {
	w := webapi.New(
		":8080",
		webapi.WithPrefix("api"),
		webapi.ResponseAsJSON(new(types.AsIsResponse)),
	)

	if err := w.RegisterServices(
		new(services.NotesAPI),
		new(services.RequestAPI),
	); err != nil {
		log.Fatal(err)
	}

	if err := w.Start(); err != nil {
		log.Fatal(err)
	}
}
