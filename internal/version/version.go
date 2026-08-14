package version

import "fmt"

// These values are replaced with -ldflags in release builds.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func String() string {
	return fmt.Sprintf("stillpoint %s (commit %s, built %s)", Version, Commit, Date)
}
