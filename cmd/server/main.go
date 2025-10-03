package main

import (
	"log"
	"os"

	"gophkeeper/internal/server/app"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	application, err := app.New(version, buildDate, logger)
	if err != nil {
		logger.Fatalf("failed to init server: %v", err)
	}
	if err := application.Run(); err != nil {
		logger.Fatalf("server stopped with error: %v", err)
	}
}
