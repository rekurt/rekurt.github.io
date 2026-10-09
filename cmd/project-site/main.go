package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rekurt/rekurt.github.io/internal/projectsite"
)

type usageError struct {
	message string
}

func (e *usageError) Error() string { return e.message }

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var commandError *usageError
		if errors.As(err, &commandError) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 1 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		printUsage(stdout)
		return nil
	}
	if len(args) == 0 {
		printUsage(stderr)
		return &usageError{message: "command is required"}
	}
	command := args[0]
	if command != "build" && command != "decorate" && command != "validate" {
		printUsage(stderr)
		return &usageError{message: fmt.Sprintf("unknown command %q", command)}
	}

	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	slug := flags.String("slug", "", "catalog product slug")
	snapshot := flags.String("snapshot", "", "catalog snapshot JSON")
	repository := flags.String("repo", "", "project repository checkout")
	output := flags.String("out", "", "static artifact directory")
	baseURL := flags.String("base-url", "", "canonical HTTPS base URL")
	if err := flags.Parse(args[1:]); err != nil {
		return &usageError{message: err.Error()}
	}
	if flags.NArg() != 0 {
		return &usageError{message: "unexpected positional arguments: " + strings.Join(flags.Args(), " ")}
	}
	required := []struct {
		name  string
		value string
	}{{"--slug", *slug}, {"--snapshot", *snapshot}, {"--repo", *repository}, {"--out", *output}, {"--base-url", *baseURL}}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			return &usageError{message: item.name + " is required"}
		}
	}
	options := projectsite.Options{
		Slug: *slug, SnapshotPath: *snapshot, Repository: *repository, Output: *output, BaseURL: *baseURL,
	}
	switch command {
	case "build":
		if _, err := projectsite.Build(options); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "built %s\n", *slug)
	case "decorate":
		if _, err := projectsite.Decorate(options); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "decorated %s\n", *slug)
	case "validate":
		if err := projectsite.Validate(options); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "validated %s\n", *slug)
	}
	return nil
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "usage: project-site <build|decorate|validate> --slug <slug> --snapshot <catalog.json> --repo <directory> --out <directory> --base-url <https-url>")
}
