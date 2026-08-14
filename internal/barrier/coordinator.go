package barrier

import (
	"context"
	"errors"
	"fmt"

	"github.com/viterineil/stillpoint/internal/snapshot"
)

// Agents combines cooperative checkpointing with a hard process fence. Every
// recovery method must be idempotent because cleanup also runs after partial
// adapter failures.
type Agents interface {
	CloseDispatchGate(ctx context.Context, barrierID string) error
	Drain(ctx context.Context, barrierID string) error
	Fence(ctx context.Context, barrierID string) error
	Unfence(ctx context.Context, barrierID string) error
	Resume(ctx context.Context, barrierID string) error
	OpenDispatchGate(ctx context.Context, barrierID string) error
}

type Coordinator struct {
	Agents    Agents
	Snapshots snapshot.Provider
}

// CapturePoint pauses writers only long enough to create a read-only snapshot.
// Conservative flags are set before each mutating call so deferred cleanup is
// attempted even if an adapter changes state and then returns an error.
func (c Coordinator) CapturePoint(ctx context.Context, barrierID string) (ref snapshot.Ref, err error) {
	if c.Agents == nil {
		return snapshot.Ref{}, errors.New("agent coordinator is required")
	}
	if c.Snapshots == nil {
		return snapshot.Ref{}, errors.New("snapshot provider is required")
	}
	if barrierID == "" {
		return snapshot.Ref{}, errors.New("barrier ID is required")
	}

	gateTouched := false
	agentsTouched := false
	fenceTouched := false
	defer func() {
		cleanupCtx := context.WithoutCancel(ctx)
		var cleanupErr error
		if fenceTouched {
			cleanupErr = errors.Join(cleanupErr, wrap("unfence agents", c.Agents.Unfence(cleanupCtx, barrierID)))
		}
		if agentsTouched {
			cleanupErr = errors.Join(cleanupErr, wrap("resume agents", c.Agents.Resume(cleanupCtx, barrierID)))
		}
		if gateTouched {
			cleanupErr = errors.Join(cleanupErr, wrap("open dispatch gate", c.Agents.OpenDispatchGate(cleanupCtx, barrierID)))
		}
		err = errors.Join(err, cleanupErr)
	}()

	gateTouched = true
	if err := c.Agents.CloseDispatchGate(ctx, barrierID); err != nil {
		return snapshot.Ref{}, fmt.Errorf("close dispatch gate: %w", err)
	}

	agentsTouched = true
	if err := c.Agents.Drain(ctx, barrierID); err != nil {
		return snapshot.Ref{}, fmt.Errorf("drain agents: %w", err)
	}

	fenceTouched = true
	if err := c.Agents.Fence(ctx, barrierID); err != nil {
		return snapshot.Ref{}, fmt.Errorf("fence agents: %w", err)
	}

	ref, err = c.Snapshots.CreateReadOnly(ctx, barrierID)
	if err != nil {
		return snapshot.Ref{}, fmt.Errorf("create read-only snapshot: %w", err)
	}
	if !ref.ReadOnly {
		return snapshot.Ref{}, errors.New("snapshot provider returned a writable snapshot")
	}
	return ref, nil
}

func wrap(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
