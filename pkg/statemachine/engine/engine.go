package engine

import (
	"fmt"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/cel"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/smctx"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/state"
	"sync"
)

type Engine struct {
	mu sync.RWMutex

	states       map[string]*state.State
	initialState string
	currentState string

	evaluator *cel.Evaluator

	// Hooks for debugging
	onTransition TransitionHook
	onStateEnter StateHook
	onStateExit  StateHook

	metadata map[string]interface{}
}

type (
	TransitionHook func(ctx *smctx.Context, from, to *state.State, transition *state.Transition)
	StateHook      func(ctx *smctx.Context, s *state.State)
)

func NewEngine(opts ...EngineOption) *Engine {
	e := &Engine{
		states:    make(map[string]*state.State),
		evaluator: cel.NewEvaluator(),
		metadata:  make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (e *Engine) AddState(s *state.State) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.states[s.ID]; exists {
		return fmt.Errorf("state %s already exists", s.ID)
	}

	e.states[s.ID] = s
	return nil
}

func (e *Engine) SetInitialState(stateID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.states[stateID]; exists {
		return fmt.Errorf("state %s already exists", stateID)
	}

	e.initialState = stateID
	if e.currentState == "" {
		e.currentState = e.initialState
	}

	return nil
}

func (e *Engine) Start(ctx *smctx.Context) error {
	e.mu.Lock()
	if e.initialState == "" {
		e.mu.Unlock()
		return fmt.Errorf("initial state is empty")
	}

	currentState := e.states[e.initialState]
	e.mu.Unlock()

	return e.executeState(ctx, currentState)
}

func (e *Engine) executeState(ctx *smctx.Context, s *state.State) error {
	if e.onStateEnter != nil {
		e.onStateEnter(ctx, s)
	}

	if err := s.Execute(ctx); err != nil {
		return err
	}

	transition := s.GetTransitions()
	for _, t := range transition {
		if t.GetCondition() == "" {
			return e.transition(ctx, s, t)
		}

		satisfied, err := e.evaluator.Evaluate(t.GetCondition(), ctx)
		if err != nil {
			return err
		}

		if satisfied {
			return e.transition(ctx, s, t)
		}
	}

	return nil
}

func (e *Engine) transition(ctx *smctx.Context, from *state.State, t *state.Transition) error {
	e.mu.Lock()
	to, exists := e.states[t.GetTargetID()]
	if !exists {
		e.mu.Unlock()
		return fmt.Errorf("transition target %s not found", t.GetTargetID())
	}
	e.mu.Unlock()

	if err := from.Exit(ctx); err != nil {
		return err
	}

	if err := t.Execute(ctx); err != nil {
		return err
	}

	if e.onTransition != nil {
		e.onTransition(ctx, from, to, t)
	}

	e.mu.Lock()
	e.currentState = to.ID
	e.mu.Unlock()

	return e.executeState(ctx, to)
}

func (e *Engine) GetCurrentState() (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.currentState == "" {
		return "", fmt.Errorf("current state is empty")
	}

	return e.currentState, nil
}

func (e *Engine) GetState(id string) (*state.State, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.states[id] == nil {
		return nil, fmt.Errorf("state %s not found", id)
	}
	return e.states[id], nil
}

type EngineOption func(*Engine)

func WithTransitionHook(hook TransitionHook) EngineOption {
	return func(e *Engine) {
		e.onTransition = hook
	}
}

func WithStateHooks(enterHook, exitHook StateHook) EngineOption {
	return func(e *Engine) {
		e.onStateEnter = enterHook
		e.onStateExit = exitHook
	}
}
