package safety

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNoSources         = errors.New("at least one source is required")
	ErrNoDestination     = errors.New("a local repository destination is required")
	ErrDestinationInside = errors.New("repository destination is inside a source")
)

// Layout is a canonicalized local backup layout.
type Layout struct {
	Sources     []string
	Destination string
}

// ValidateLocalLayout resolves symlinks as far as the filesystem allows and
// refuses a local repository nested inside any source. Nonexistent destination
// leaf paths are accepted so a repository can be initialized later.
func ValidateLocalLayout(sources []string, destination string) (Layout, error) {
	if len(sources) == 0 {
		return Layout{}, ErrNoSources
	}
	if strings.TrimSpace(destination) == "" {
		return Layout{}, ErrNoDestination
	}

	dest, err := canonical(destination)
	if err != nil {
		return Layout{}, fmt.Errorf("resolve repository destination: %w", err)
	}

	result := Layout{Destination: dest}
	seen := make(map[string]struct{}, len(sources))
	for _, source := range sources {
		if strings.TrimSpace(source) == "" {
			return Layout{}, errors.New("source path may not be empty")
		}
		resolved, err := canonical(source)
		if err != nil {
			return Layout{}, fmt.Errorf("resolve source %q: %w", source, err)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return Layout{}, fmt.Errorf("inspect source %q: %w", source, err)
		}
		if !info.IsDir() {
			return Layout{}, fmt.Errorf("source %q is not a directory", source)
		}
		if _, duplicate := seen[resolved]; duplicate {
			return Layout{}, fmt.Errorf("source %q is selected more than once", resolved)
		}
		seen[resolved] = struct{}{}
		if contains(resolved, dest) {
			return Layout{}, fmt.Errorf("%w: %s is within %s", ErrDestinationInside, dest, resolved)
		}
		result.Sources = append(result.Sources, resolved)
	}

	return result, nil
}

func contains(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// canonical resolves an existing prefix and preserves nonexistent path
// components. This catches symlink-based nesting even before a repository has
// been created.
func canonical(path string) (string, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}

	current := abs
	var missing []string
	for {
		_, statErr := os.Lstat(current)
		if statErr == nil {
			break
		}
		if !os.IsNotExist(statErr) {
			return "", statErr
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", statErr
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}

	resolved, err := filepath.EvalSymlinks(current)
	if err != nil {
		return "", err
	}
	for i := len(missing) - 1; i >= 0; i-- {
		resolved = filepath.Join(resolved, missing[i])
	}
	return filepath.Clean(resolved), nil
}
