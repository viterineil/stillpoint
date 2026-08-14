package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanDirectMode(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	repository := filepath.Join(root, "repository")
	var stdout, stderr bytes.Buffer
	exit := Run([]string{"plan", "--source", source, "--repository", repository, "--json"}, &stdout, &stderr)
	if exit != exitOK {
		t.Fatalf("exit = %d, stderr = %s", exit, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"executable": false`) {
		t.Fatalf("output does not contain safety gate: %s", stdout.String())
	}
}

func TestBackupIsSafetyGated(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exit := Run([]string{"backup"}, &stdout, &stderr); exit != exitUnavailable {
		t.Fatalf("exit = %d, want %d", exit, exitUnavailable)
	}
	if !strings.Contains(stderr.String(), "intentionally disabled") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestUnknownCommandIsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exit := Run([]string{"nope"}, &stdout, &stderr); exit != exitUsage {
		t.Fatalf("exit = %d, want %d", exit, exitUsage)
	}
}
