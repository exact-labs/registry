package main

import (
	"log"

	"just/pkg/helpers"
	"just/pkg/routes"
	"just/pkg/templates"

	"github.com/pocketbase/pocketbase"
)

func main() {
	_, isUsingGoRun := helpers.InspectRuntime()

	app := pocketbase.NewWithConfig(&pocketbase.Config{
		DefaultDataDir: "packages",
		DefaultDebug:   isUsingGoRun,
	})

	if err := routes.Router(app); err != nil {
		log.Fatal(err)
	}

	if err := templates.Copy(); err != nil {
		log.Fatal(err)
	}

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
