package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/pivotal-cf/pivnet-resource/ui"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string
)

func main() {
	if version == "" {
		version = "dev"
	}

	color.NoColor = false

	logWriter := os.Stderr
	uiPrinter := ui.NewUIPrinter(logWriter)

	logger := log.New(logWriter, "", log.LstdFlags)

	logger.Printf("PivNet Product Stemcell Resource version: %s", version)

	uiPrinter.PrintErrorln(fmt.Errorf("not implemented"))
	os.Exit(1)
}