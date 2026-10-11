package main

import (
	"fmt"
	"os"

	"labs/gira-g/internal/domain"
)

var (
	version = "0.1.0-dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version") {
		fmt.Printf("gira %s (commit: %s, date: %s)\n", version, commit, date)
		return
	}

	mm := domain.NewMacroMap("gira-harness")
	out, err := mm.CanonicalJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing macro map: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Gira-G Architecture Harness (%s)\n", version)
	fmt.Printf("Target System initialized: %s\n", mm.TargetSystem)
	_ = out
}
