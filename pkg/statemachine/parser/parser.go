package parser

import (
	"encoding/json"
	"fmt"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/engine"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/FinnTew/FinnFlow/pkg/statemachine/action"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/state"
)

type StateMachineConfig struct {
	Version      string                 `json:"version" yaml:"version"`
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description" yaml:"description"`
	InitialState string                 `json:"initialState" yaml:"initialState"`
	States       map[string]StateConfig `json:"states" yaml:"states"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type StateConfig struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Type        string `json:"type" yaml:"type"`
	IsFinal     bool   `json:"isFinal" yaml:"isFinal"`

	EntryActions []ActionConfig         `json:"entryActions,omitempty" yaml:"entryActions,omitempty"`
	ExitActions  []ActionConfig         `json:"exitActions,omitempty" yaml:"exitActions,omitempty"`
	Transitions  []TransitionConfig     `json:"transitions,omitempty" yaml:"transitions,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type TransitionConfig struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description" yaml:"description"`
	Target      string                 `json:"target" yaml:"target"`
	Condition   string                 `json:"condition,omitempty" yaml:"condition,omitempty"`
	Actions     []ActionConfig         `json:"actions,omitempty" yaml:"actions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type ActionConfig struct {
	Type       string                 `json:"type" yaml:"type"`
	Name       string                 `json:"name" yaml:"name"`
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

type Parser struct {
	actionRegistry ActionRegistry
}

type ActionRegistry interface {
	CreateAction(config ActionConfig) (action.Action, error)
}

func NewParser(registry ActionRegistry) *Parser {
	return &Parser{
		actionRegistry: registry,
	}
}

func (p *Parser) ParseJSON(reader io.Reader) (*StateMachineConfig, error) {
	var config StateMachineConfig
	if err := json.NewDecoder(reader).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &config, nil
}

func (p *Parser) ParseYAML(reader io.Reader) (*StateMachineConfig, error) {
	var config StateMachineConfig
	if err := yaml.NewDecoder(reader).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return &config, nil
}

func (p *Parser) BuildStateMachine(config *StateMachineConfig) (*engine.Engine, error) {
	e := engine.NewEngine()

	for id, stateConfig := range config.States {
		s, err := p.buildState(id, stateConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build state %s: %w", id, err)
		}

		if err := e.AddState(s); err != nil {
			return nil, fmt.Errorf("failed to add state %s: %w", id, err)
		}
	}

	for id, stateConfig := range config.States {
		if err := p.addTransitions(e, id, stateConfig); err != nil {
			return nil, fmt.Errorf("failed to add transitions for state %s: %w", id, err)
		}
	}

	if err := e.SetInitialState(config.InitialState); err != nil {
		return nil, fmt.Errorf("failed to set initial state: %w", err)
	}

	return e, nil
}

func (p *Parser) buildState(id string, config StateConfig) (*state.State, error) {
	entryActions := make([]action.Action, 0, len(config.EntryActions))
	for _, actionConfig := range config.EntryActions {
		act, err := p.actionRegistry.CreateAction(actionConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create entry action: %w", err)
		}
		entryActions = append(entryActions, act)
	}

	exitActions := make([]action.Action, 0, len(config.ExitActions))
	for _, actionConfig := range config.ExitActions {
		act, err := p.actionRegistry.CreateAction(actionConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create exit action: %w", err)
		}
		exitActions = append(exitActions, act)
	}

	return state.NewState(id,
		state.WithName(config.Name),
		state.WithDescription(config.Description),
		state.WithEntryActions(entryActions...),
		state.WithExitActions(exitActions...),
		state.WithFinal(config.IsFinal),
	), nil
}

func (p *Parser) addTransitions(engine *engine.Engine, sourceID string, config StateConfig) error {
	for _, transConfig := range config.Transitions {
		transActions := make([]action.Action, 0, len(transConfig.Actions))
		for _, actionConfig := range transConfig.Actions {
			act, err := p.actionRegistry.CreateAction(actionConfig)
			if err != nil {
				return fmt.Errorf("failed to create transition action: %w", err)
			}
			transActions = append(transActions, act)
		}

		transition := state.NewTransition(
			sourceID,
			transConfig.Target,
			state.WithTransitionName(transConfig.Name),
			state.WithTransitionDescription(transConfig.Description),
			state.WithTransitionCondition(transConfig.Condition),
			state.WithTransitionActions(transActions...),
		)

		sourceState, err := engine.GetState(sourceID)
		if err != nil {
			return err
		}

		sourceState.AddTransition(transition)
	}

	return nil
}
