package parser

import (
	"fmt"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/action"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/smctx"
	"sync"
)

type DefaultRegistry struct {
	mu       sync.RWMutex
	creators map[string]ActionCreator
}

type ActionCreator func(config map[string]interface{}) (action.Action, error)

func NewDefaultRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		creators: make(map[string]ActionCreator),
	}
}

func (r *DefaultRegistry) Register(actionType string, creator ActionCreator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.creators[actionType] = creator
}

func (r *DefaultRegistry) CreateAction(config ActionConfig) (action.Action, error) {
	r.mu.RLock()
	creator, exists := r.creators[config.Type]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown action type: %s", config.Type)
	}

	return creator(config.Parameters)
}

func (r *DefaultRegistry) RegisterCommonActions() {
	r.Register("log", func(params map[string]interface{}) (action.Action, error) {
		message, _ := params["message"].(string)
		return action.ActionFunc(func(ctx *smctx.Context) error {
			// TODO: Implement logging logic
			fmt.Println(message)
			return nil
		}), nil
	})

	r.Register("delay", func(params map[string]interface{}) (action.Action, error) {
		duration, _ := params["duration"].(string)
		return action.ActionFunc(func(ctx *smctx.Context) error {
			// TODO: Parse duration and implement delay logic
			fmt.Println(duration)
			return nil
		}), nil
	})

	// Add more common actions as needed
}
