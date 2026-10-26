package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/manipulator"
)

func runReadSymbol(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.SetOutput(stderr)

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) < 2 {
		return fmt.Errorf("missing arguments. Usage: gira read <file> <symbol>")
	}

	filePath := allPositionals[0]
	symbolQuery := allPositionals[1]

	snippet, _, err := manipulator.ReadSymbol(filePath, symbolQuery)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, snippet)
	return nil
}

func runWriteSymbol(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("write", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var body string
	fs.StringVar(&body, "body", "", "New content for the symbol (or read from stdin if omitted)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) < 2 {
		return fmt.Errorf("missing arguments. Usage: gira write <file> <symbol> [--body <content>]")
	}

	filePath := allPositionals[0]
	symbolQuery := allPositionals[1]

	content := body
	if content == "" || content == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed reading from stdin: %w", err)
		}
		content = string(data)
	}

	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("cannot write empty symbol body")
	}

	if err := manipulator.WriteSymbol(filePath, symbolQuery, content); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "successfully updated symbol %q in %s\n", symbolQuery, filePath)
	return nil
}

func runAddSymbol(args []string, stdout io.Writer, stderr io.Writer) error {
	positionals, flagArgs := splitArgs(args)

	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var body string
	fs.StringVar(&body, "body", "", "New member content (or read from stdin if omitted)")

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	allPositionals := append(positionals, fs.Args()...)
	if len(allPositionals) < 2 {
		return fmt.Errorf("missing arguments. Usage: gira add <file> <container> [--body <content>]")
	}

	filePath := allPositionals[0]
	containerQuery := allPositionals[1]

	content := body
	if content == "" || content == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed reading from stdin: %w", err)
		}
		content = string(data)
	}

	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("cannot add empty member body")
	}

	if err := manipulator.AddMember(filePath, containerQuery, content); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "successfully added member to %q in %s\n", containerQuery, filePath)
	return nil
}
