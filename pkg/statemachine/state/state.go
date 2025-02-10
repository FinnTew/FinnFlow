package state

import (
	"github.com/FinnTew/FinnFlow/pkg/statemachine/action"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/smctx"
	"sync"
)

type State struct {
	mu sync.RWMutex

	ID          string
	Name        string
	Description string

	entryActions []action.Action
	exitActions  []action.Action
	transitions  []*Transition

	metadata map[string]interface{}
	isFinal  bool
}

func NewState(id string, opts ...StateOption) *State {
	s := &State{
		ID:          id,
		metadata:    make(map[string]interface{}),
		transitions: make([]*Transition, 0),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *State) AddTransition(t *Transition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transitions = append(s.transitions, t)
}

func (s *State) GetTransitions() []*Transition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Transition, len(s.transitions))
	copy(result, s.transitions)
	return result
}

func (s *State) Execute(ctx *smctx.Context) error {
	s.mu.RLock()
	entryActions := s.entryActions
	s.mu.RUnlock()

	for _, act := range entryActions {
		if err := act.Execute(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) Exit(ctx *smctx.Context) error {
	s.mu.RLock()
	exitActions := s.exitActions
	s.mu.RUnlock()

	for _, act := range exitActions {
		if err := act.Execute(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) IsFinal() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isFinal
}

type StateOption func(*State)

func WithName(name string) StateOption {
	return func(s *State) {
		s.Name = name
	}
}

func WithDescription(desc string) StateOption {
	return func(s *State) {
		s.Description = desc
	}
}

func WithEntryActions(actions ...action.Action) StateOption {
	return func(s *State) {
		s.entryActions = actions
	}
}

func WithExitActions(actions ...action.Action) StateOption {
	return func(s *State) {
		s.exitActions = actions
	}
}

func WithFinal(isFinal bool) StateOption {
	return func(s *State) {
		s.isFinal = isFinal
	}
}
