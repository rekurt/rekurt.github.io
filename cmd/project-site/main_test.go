package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBuildValidateAndDecorate(t *testing.T) {
	repository, err := filepath.Abs("../../internal/projectsite/testdata/repository")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := filepath.Abs("../../internal/projectsite/testdata/snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "generated")
	base := []string{"--slug", "git-barber", "--snapshot", snapshot, "--repo", repository, "--out", output, "--base-url", "https://rekurt.github.io/git-barber/"}
	var stdout, stderr bytes.Buffer
	if err := run(append([]string{"build"}, base...), &stdout, &stderr); err != nil {
		t.Fatalf("build: %v / %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `built git-barber`) {
		t.Fatalf("build output = %q", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if err := run(append([]string{"validate"}, base...), &stdout, &stderr); err != nil {
		t.Fatalf("validate: %v / %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `validated git-barber`) {
		t.Fatalf("validate output = %q", stdout.String())
	}

	decorated := t.TempDir()
	index, err := os.ReadFile("../../internal/projectsite/testdata/existing/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(decorated, "index.html"), index, 0o644); err != nil {
		t.Fatal(err)
	}
	decorateArgs := append([]string{"decorate"}, append(base[:len(base)-4], "--out", decorated, "--base-url", "https://rekurt.github.io/git-barber/")...)
	stdout.Reset()
	stderr.Reset()
	if err := run(decorateArgs, &stdout, &stderr); err != nil {
		t.Fatalf("decorate: %v / %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `decorated git-barber`) {
		t.Fatalf("decorate output = %q", stdout.String())
	}
}

func TestRunReportsUsageErrors(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{args: nil, want: "usage:"},
		{args: []string{"unknown"}, want: "unknown command"},
		{args: []string{"build", "--slug", "git-barber"}, want: "--snapshot is required"},
	}
	for _, tt := range tests {
		var stdout, stderr bytes.Buffer
		err := run(tt.args, &stdout, &stderr)
		if err == nil || !strings.Contains(err.Error()+stderr.String(), tt.want) {
			t.Fatalf("run(%v) error/stderr = %v / %q, want %q", tt.args, err, stderr.String(), tt.want)
		}
	}
}

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"help"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "project-site <build|decorate|validate>") || stderr.Len() != 0 {
		t.Fatalf("stdout/stderr = %q / %q", stdout.String(), stderr.String())
	}
}
