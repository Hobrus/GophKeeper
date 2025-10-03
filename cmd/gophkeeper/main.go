package main

import (
	"fmt"
	"os"

	"gophkeeper/internal/client/cmd"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	root := cmd.NewRootCmd(version, buildDate)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
