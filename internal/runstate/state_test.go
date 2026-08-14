package runstate

import "testing"

func TestMachineHappyPath(t *testing.T) {
	m := New()
	states := []State{Draining, Quiescent, Fenced, FSSnapshotted, Resumed, ComponentsBackedUp, LocalComplete, Replicated, Verified}
	for _, state := range states {
		if err := m.Transition(state); err != nil {
			t.Fatalf("Transition(%s) error = %v", state, err)
		}
	}
	if got := m.State(); got != Verified {
		t.Fatalf("state = %s, want %s", got, Verified)
	}
}

func TestMachineRejectsSkippedState(t *testing.T) {
	m := New()
	if err := m.Transition(Fenced); err == nil {
		t.Fatal("Transition(Fenced) unexpectedly succeeded")
	}
}

func TestMachineCanAbortBeforeLocalCompletion(t *testing.T) {
	m := New()
	if err := m.Transition(Draining); err != nil {
		t.Fatal(err)
	}
	if err := m.Transition(Aborted); err != nil {
		t.Fatal(err)
	}
	if got := m.State(); got != Aborted {
		t.Fatalf("state = %s, want %s", got, Aborted)
	}
}
