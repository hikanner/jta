package main

import (
	"fmt"
	"os"

	"github.com/hikanner/jta/internal/cli"
)

var (
	// Version information (will be set by build tools)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Pass version information to CLI
	cli.Version = version
	cli.Commit = commit
	cli.Date = date

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
