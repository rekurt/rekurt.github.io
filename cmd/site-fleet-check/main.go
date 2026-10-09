package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/sitefleet"
)

type checker func(context.Context, sitefleet.Options) ([]sitefleet.Result, error)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, sitefleet.Check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer, check checker) error {
	flags := flag.NewFlagSet("site-fleet-check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	snapshot := flags.String("snapshot", "", "catalog snapshot JSON")
	timeout := flags.Duration("timeout", 5*time.Minute, "whole-fleet timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	if strings.TrimSpace(*snapshot) == "" {
		return fmt.Errorf("--snapshot is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	results, checkErr := check(ctx, sitefleet.Options{SnapshotPath: *snapshot})
	for _, result := range results {
		status := "failed"
		if result.Verified {
			status = "verified"
		}
		fmt.Fprintf(stdout, "%s\t%s\t%d\t%s\t%s\n", result.Slug, result.Mode, len(result.Checked), status, result.URL)
	}
	return checkErr
}
