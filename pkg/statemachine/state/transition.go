package state

import (
	"github.com/FinnTew/FinnFlow/pkg/statemachine/action"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/smctx"
)

type Transition struct {
	ID          string
	Name        string
	Description string

	sourceID  string
	targetID  string
	condition string
	actions   []action.Action

	metadata map[string]interface{}
}

func NewTransition(sourceID, targetID string, opts ...TransitionOption) *Transition {
	t := &Transition{
		sourceID: sourceID,
		targetID: targetID,
		metadata: make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *Transition) Execute(ctx *smctx.Context) error {
	for _, act := range t.actions {
		if err := act.Execute(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transition) GetCondition() string {
	return t.condition
}

func (t *Transition) GetTargetID() string {
	return t.targetID
}

type TransitionOption func(*Transition)

func WithTransitionName(name string) TransitionOption {
	return func(t *Transition) {
		t.Name = name
	}
}

func WithTransitionDescription(description string) TransitionOption {
	return func(t *Transition) {
		t.Description = description
	}
}

func WithTransitionCondition(condition string) TransitionOption {
	return func(t *Transition) {
		t.condition = condition
	}
}

func WithTransitionActions(actions ...action.Action) TransitionOption {
	return func(t *Transition) {
		t.actions = actions
	}
}
