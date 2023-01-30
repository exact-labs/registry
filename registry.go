package main

import (
   "os"
	"log"

	"registry/pkg/helpers"
	"registry/pkg/routes"
	"registry/pkg/templates"

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
   
   version, err := helpers.GetJustVersion(); 
   if err != nil {
      log.Fatal(err)
   } else {
      os.Setenv("JUST_VERSION", version)
   }
   
	if err := templates.Copy(); err != nil {
		log.Fatal(err)
	}

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
