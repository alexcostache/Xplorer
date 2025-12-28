package main

import (
	"flag"
	"log"

	"github.com/alexcostache/Xplorer/internal/app"
)

func main() {
	// Parse command line flags
	debugFlag := flag.Bool("debug", false, "Enable debug logging to /tmp/xp_debug.log")
	flag.Parse()
	
	application := app.New()
	
	// Enable debug mode if flag is set
	if *debugFlag {
		application.EnableDebug()
	}
	
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}

// Made with Bob
