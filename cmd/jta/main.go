package main

import (
	"fmt"
	"os"

	"github.com/hikanner/jta/internal/cli"
)

var (
	// Version information (will be set by GoReleaser)
	version = "dev"     //nolint:unused // Will be set by build tools
	commit  = "none"    //nolint:unused // Will be set by build tools
	date    = "unknown" //nolint:unused // Will be set by build tools
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
