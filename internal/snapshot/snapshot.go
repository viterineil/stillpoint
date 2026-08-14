package snapshot

import (
	"context"
	"time"
)

type Ref struct {
	Provider  string    `json:"provider"`
	Path      string    `json:"path"`
	ReadOnly  bool      `json:"read_only"`
	CreatedAt time.Time `json:"created_at"`
}

// Provider creates the immutable filesystem instant from which the longer
// Restic backup can safely run after agents resume.
type Provider interface {
	CreateReadOnly(ctx context.Context, backupSetID string) (Ref, error)
}
