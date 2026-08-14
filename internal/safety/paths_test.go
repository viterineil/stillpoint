package safety

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateLocalLayoutAcceptsSeparateDestination(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "backup", "repo")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}

	layout, err := ValidateLocalLayout([]string{source}, destination)
	if err != nil {
		t.Fatalf("ValidateLocalLayout() error = %v", err)
	}
	if layout.Destination != destination {
		t.Fatalf("destination = %q, want %q", layout.Destination, destination)
	}
}

func TestValidateLocalLayoutRejectsNestedDestination(t *testing.T) {
	source := t.TempDir()
	destination := filepath.Join(source, ".backups", "restic")

	_, err := ValidateLocalLayout([]string{source}, destination)
	if !errors.Is(err, ErrDestinationInside) {
		t.Fatalf("error = %v, want ErrDestinationInside", err)
	}
}

func TestValidateLocalLayoutResolvesSymlink(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	link := filepath.Join(root, "linked-source")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, link); err != nil {
		t.Fatal(err)
	}

	_, err := ValidateLocalLayout([]string{link}, filepath.Join(source, "repo"))
	if !errors.Is(err, ErrDestinationInside) {
		t.Fatalf("error = %v, want ErrDestinationInside", err)
	}
}
