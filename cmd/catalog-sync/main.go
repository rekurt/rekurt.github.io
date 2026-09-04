package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	stdsync "sync"
	"time"

	"github.com/rekurt/rekurt.github.io/internal/catalog"
	"github.com/rekurt/rekurt.github.io/internal/githubapi"
	catalogsync "github.com/rekurt/rekurt.github.io/internal/sync"
)

type usageError struct {
	message string
}

func (e *usageError) Error() string { return e.message }

type repositoryClient interface {
	ListOwnedPublic(context.Context, string) ([]catalog.Repository, error)
	Enrich(context.Context, catalog.Repository, bool) (catalog.Repository, error)
}

func main() {
	err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr)
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	var commandError *usageError
	if errors.As(err, &commandError) {
		os.Exit(2)
	}
	os.Exit(1)
}

func run(ctx context.Context, args []string, getenv func(string) string, out, errOut io.Writer) error {
	if len(args) == 0 || (args[0] != "sync" && args[0] != "check") {
		printUsage(errOut)
		command := ""
		if len(args) > 0 {
			command = args[0]
		}
		return &usageError{message: fmt.Sprintf("unknown command %q", command)}
	}
	command := args[0]
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(errOut)
	manifestPath := flags.String("manifest", "catalog/projects.yaml", "curated project manifest")
	snapshotPath := flags.String("snapshot", "site/src/data/generated/catalog.json", "generated catalog snapshot")
	auditPath := flags.String("audit", "docs/repository-audit.md", "generated repository audit")
	if err := flags.Parse(args[1:]); err != nil {
		printUsage(errOut)
		return &usageError{message: err.Error()}
	}
	if flags.NArg() != 0 {
		printUsage(errOut)
		return &usageError{message: "unexpected positional arguments: " + strings.Join(flags.Args(), " ")}
	}

	manifest, err := catalog.LoadManifest(*manifestPath)
	if err != nil {
		return err
	}
	if err := catalog.ValidateManifest(manifest); err != nil {
		return fmt.Errorf("validate manifest: %w", err)
	}

	client := githubapi.New(getenv("GITHUB_API_URL"), getenv("GITHUB_TOKEN"), &http.Client{Timeout: 30 * time.Second})
	repositories, err := client.ListOwnedPublic(ctx, manifest.Owner)
	if err != nil {
		return fmt.Errorf("list public repositories: %w", err)
	}
	primary := make(map[string]struct{}, len(manifest.Products))
	for _, product := range manifest.Products {
		primary[strings.ToLower(product.PrimaryRepo)] = struct{}{}
	}
	if err := enrichAll(ctx, client, repositories, primary); err != nil {
		return err
	}

	candidate, err := catalogsync.Build(manifest, repositories, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("build catalog: %w", err)
	}
	if command == "check" {
		changed, err := catalogsync.SnapshotChanged(*snapshotPath, candidate)
		if err != nil {
			return fmt.Errorf("compare catalog snapshot: %w", err)
		}
		if changed {
			return errors.New("catalog snapshot is out of date")
		}
		existing, err := catalogsync.ReadSnapshot(*snapshotPath)
		if err != nil {
			return err
		}
		actualAudit, err := os.ReadFile(*auditPath)
		if err != nil {
			return fmt.Errorf("read audit report: %w", err)
		}
		if !bytes.Equal(actualAudit, catalogsync.RenderAudit(existing)) {
			return errors.New("repository audit is out of date")
		}
		fmt.Fprintf(out, "catalog is current: %d products, %d repositories\n", len(existing.Products), len(existing.Repositories))
		return nil
	}

	snapshotChanged, err := catalogsync.WriteSnapshot(*snapshotPath, candidate)
	if err != nil {
		return fmt.Errorf("write catalog snapshot: %w", err)
	}
	effective, err := catalogsync.ReadSnapshot(*snapshotPath)
	if err != nil {
		return err
	}
	auditChanged, err := catalogsync.WriteBytesIfChanged(*auditPath, catalogsync.RenderAudit(effective))
	if err != nil {
		return fmt.Errorf("write repository audit: %w", err)
	}
	fmt.Fprintf(out, "catalog synced: %d products, %d repositories (snapshot_changed=%t audit_changed=%t)\n",
		len(effective.Products), len(effective.Repositories), snapshotChanged, auditChanged)
	return nil
}

func enrichAll(ctx context.Context, client repositoryClient, repositories []catalog.Repository, primary map[string]struct{}) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan int)
	var workers stdsync.WaitGroup
	var firstErr error
	var errOnce stdsync.Once
	worker := func() {
		defer workers.Done()
		for index := range jobs {
			_, withReadme := primary[strings.ToLower(repositories[index].NameWithOwner)]
			enriched, err := client.Enrich(ctx, repositories[index], withReadme)
			if err != nil {
				errOnce.Do(func() {
					firstErr = fmt.Errorf("enrich %s: %w", repositories[index].NameWithOwner, err)
					cancel()
				})
				continue
			}
			repositories[index] = enriched
		}
	}
	workerCount := 4
	if len(repositories) < workerCount {
		workerCount = len(repositories)
	}
	workers.Add(workerCount)
	for range workerCount {
		go worker()
	}

sendLoop:
	for index := range repositories {
		select {
		case <-ctx.Done():
			break sendLoop
		case jobs <- index:
		}
	}
	close(jobs)
	workers.Wait()
	return firstErr
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "usage: catalog-sync <sync|check> [--manifest path] [--snapshot path] [--audit path]")
}
