package action

import "github.com/FinnTew/FinnFlow/pkg/statemachine/smctx"

type Action interface {
	Execute(ctx *smctx.Context) error
	Name() string
}

type ActionFunc func(ctx *smctx.Context) error

func (f ActionFunc) Execute(ctx *smctx.Context) error {
	return f(ctx)
}

func (f ActionFunc) Name() string {
	return "action"
}

type CompositeAction struct {
	name    string
	actions []Action
}

func NewCompositeAction(name string, actions ...Action) *CompositeAction {
	return &CompositeAction{
		name:    name,
		actions: actions,
	}
}

func (ca *CompositeAction) Execute(ctx *smctx.Context) error {
	for _, action := range ca.actions {
		if err := action.Execute(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (ca *CompositeAction) Name() string {
	return ca.name
}
