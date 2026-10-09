package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rekurt/rekurt.github.io/internal/sitefleet"
)

func TestRunReportsVerifiedFleet(t *testing.T) {
	var stdout, stderr bytes.Buffer
	check := func(context.Context, sitefleet.Options) ([]sitefleet.Result, error) {
		return []sitefleet.Result{{Slug: "alpha", URL: "https://rekurt.github.io/alpha/", Mode: "build", Checked: []string{"root"}, Verified: true}}, nil
	}
	if err := run([]string{"--snapshot", "catalog.json"}, &stdout, &stderr, check); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "alpha\tbuild\t1\tverified") || stderr.Len() != 0 {
		t.Fatalf("stdout/stderr = %q / %q", stdout.String(), stderr.String())
	}
}

func TestRunPreservesCheckerFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	want := errors.New("broken site")
	check := func(context.Context, sitefleet.Options) ([]sitefleet.Result, error) {
		return []sitefleet.Result{{Slug: "alpha", URL: "https://rekurt.github.io/alpha/", Mode: "build", Checked: []string{"root"}}}, want
	}
	err := run([]string{"--snapshot", "catalog.json"}, &stdout, &stderr, check)
	if !errors.Is(err, want) || !strings.Contains(stdout.String(), "failed") {
		t.Fatalf("error/stdout = %v / %q", err, stdout.String())
	}
}

func TestRunRequiresSnapshot(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(nil, &stdout, &stderr, sitefleet.Check)
	if err == nil || !strings.Contains(err.Error(), "--snapshot is required") {
		t.Fatalf("error = %v", err)
	}
}
