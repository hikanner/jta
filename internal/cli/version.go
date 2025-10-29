package cli

import (
	"fmt"
)

var (
	// Version information - set from main package
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// PrintVersion prints the version information
func PrintVersion() {
	fmt.Printf("jta version %s\n", Version)
	fmt.Printf("Git commit: %s\n", Commit)
	fmt.Printf("Built:      %s\n", Date)
}
