package barrier

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/viterineil/stillpoint/internal/snapshot"
)

type fakeAgents struct {
	calls  []string
	failAt string
}

func (f *fakeAgents) call(name string) error {
	f.calls = append(f.calls, name)
	if name == f.failAt {
		return errors.New("injected failure")
	}
	return nil
}

func (f *fakeAgents) CloseDispatchGate(context.Context, string) error { return f.call("close") }
func (f *fakeAgents) Drain(context.Context, string) error             { return f.call("drain") }
func (f *fakeAgents) Fence(context.Context, string) error             { return f.call("fence") }
func (f *fakeAgents) Unfence(context.Context, string) error           { return f.call("unfence") }
func (f *fakeAgents) Resume(context.Context, string) error            { return f.call("resume") }
func (f *fakeAgents) OpenDispatchGate(context.Context, string) error  { return f.call("open") }

type fakeSnapshotter struct {
	err error
}

func (f fakeSnapshotter) CreateReadOnly(context.Context, string) (snapshot.Ref, error) {
	if f.err != nil {
		return snapshot.Ref{}, f.err
	}
	return snapshot.Ref{Provider: "fake", Path: "/snapshot", ReadOnly: true}, nil
}

func TestCapturePointResumesAfterSuccess(t *testing.T) {
	agents := &fakeAgents{}
	coordinator := Coordinator{Agents: agents, Snapshots: fakeSnapshotter{}}
	if _, err := coordinator.CapturePoint(context.Background(), "set-1"); err != nil {
		t.Fatalf("CapturePoint() error = %v", err)
	}
	want := []string{"close", "drain", "fence", "unfence", "resume", "open"}
	if !reflect.DeepEqual(agents.calls, want) {
		t.Fatalf("calls = %v, want %v", agents.calls, want)
	}
}

func TestCapturePointCleansUpAfterFenceFailure(t *testing.T) {
	agents := &fakeAgents{failAt: "fence"}
	coordinator := Coordinator{Agents: agents, Snapshots: fakeSnapshotter{}}
	if _, err := coordinator.CapturePoint(context.Background(), "set-1"); err == nil {
		t.Fatal("CapturePoint() unexpectedly succeeded")
	}
	want := []string{"close", "drain", "fence", "unfence", "resume", "open"}
	if !reflect.DeepEqual(agents.calls, want) {
		t.Fatalf("calls = %v, want %v", agents.calls, want)
	}
}

func TestCapturePointCleansUpAfterSnapshotFailure(t *testing.T) {
	agents := &fakeAgents{}
	coordinator := Coordinator{Agents: agents, Snapshots: fakeSnapshotter{err: errors.New("snapshot failed")}}
	if _, err := coordinator.CapturePoint(context.Background(), "set-1"); err == nil {
		t.Fatal("CapturePoint() unexpectedly succeeded")
	}
	want := []string{"close", "drain", "fence", "unfence", "resume", "open"}
	if !reflect.DeepEqual(agents.calls, want) {
		t.Fatalf("calls = %v, want %v", agents.calls, want)
	}
}
