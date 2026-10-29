package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
)

// capabilitiesDoc is the context-provider protocol declaration answered by
// `gira capabilities --json` (PLAN-0009 TASK-03). Only capabilities that
// already exist as commands in Run's switch are listed.
type capabilitiesDoc struct {
	SchemaVersion int      `json:"schema_version"`
	Capabilities  []string `json:"capabilities"`
	Extras        []string `json:"extras"`
}

func capabilitiesPayload() capabilitiesDoc {
	return capabilitiesDoc{
		SchemaVersion: 1,
		Capabilities:  []string{"outline", "slice", "impact", "read"},
		Extras:        []string{"gira:tour", "gira:docs"},
	}
}

func runCapabilities(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("capabilities", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var asJSON bool
	fs.BoolVar(&asJSON, "json", false, "Output capabilities as structured JSON")

	if err := fs.Parse(args); err != nil {
		return err
	}

	doc := capabilitiesPayload()

	if asJSON {
		data, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "%s\n", data)
		return err
	}

	fmt.Fprintf(stdout, "Gira-G context provider capabilities (schema v%d)\n\n", doc.SchemaVersion)
	fmt.Fprint(stdout, "Capabilities:\n")
	for _, c := range doc.Capabilities {
		fmt.Fprintf(stdout, "  %s\n", c)
	}
	fmt.Fprint(stdout, "Extras:\n")
	for _, e := range doc.Extras {
		fmt.Fprintf(stdout, "  %s\n", e)
	}
	return nil
}
