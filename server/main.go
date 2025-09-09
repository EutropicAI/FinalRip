package main

import (
	"log"
	"os"

	"github.com/EutropicAI/FinalRip/common/version"
	"github.com/EutropicAI/FinalRip/server/cmd"
)

func main() {
	app := cmd.NewApp()
	app.Name = "FinalRip"
	app.Usage = "FinalRip API Sever"
	app.Description = "a distributed video processing tool"
	app.Version = version.FINALRIP_VERSION

	err := app.Run(os.Args)
	if err != nil {
		log.Printf("Failed to run with %s: %v\\n", os.Args, err)
	}
}
