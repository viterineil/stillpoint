package runstate

import (
	"fmt"
	"sync"
	"time"
)

type State string

const (
	Planned            State = "PLANNED"
	Draining           State = "DRAINING"
	Quiescent          State = "QUIESCENT"
	Fenced             State = "FENCED"
	FSSnapshotted      State = "FS_SNAPSHOTTED"
	Resumed            State = "RESUMED"
	ComponentsBackedUp State = "COMPONENTS_BACKED_UP"
	LocalComplete      State = "LOCAL_COMPLETE"
	Replicated         State = "REPLICATED"
	Verified           State = "VERIFIED"
	Aborted            State = "ABORTED"
)

var next = map[State]State{
	Planned:            Draining,
	Draining:           Quiescent,
	Quiescent:          Fenced,
	Fenced:             FSSnapshotted,
	FSSnapshotted:      Resumed,
	Resumed:            ComponentsBackedUp,
	ComponentsBackedUp: LocalComplete,
	LocalComplete:      Replicated,
	Replicated:         Verified,
}

type Transition struct {
	From State     `json:"from"`
	To   State     `json:"to"`
	At   time.Time `json:"at"`
}

// Machine models the monotonic run lifecycle. Durable persistence is a storage
// concern; this type keeps transition rules centralized and testable.
type Machine struct {
	mu      sync.Mutex
	state   State
	history []Transition
	now     func() time.Time
}

func New() *Machine {
	return &Machine{state: Planned, now: time.Now}
}

func (m *Machine) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

func (m *Machine) History() []Transition {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Transition(nil), m.history...)
}

func (m *Machine) Transition(to State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if to == Aborted && m.state != LocalComplete && m.state != Replicated && m.state != Verified && m.state != Aborted {
		m.record(to)
		return nil
	}
	if expected, ok := next[m.state]; !ok || expected != to {
		return fmt.Errorf("invalid backup state transition %s -> %s", m.state, to)
	}
	m.record(to)
	return nil
}

func (m *Machine) record(to State) {
	m.history = append(m.history, Transition{From: m.state, To: to, At: m.now().UTC()})
	m.state = to
}
