package cel

import (
	"fmt"
	"github.com/FinnTew/FinnFlow/pkg/statemachine/smctx"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"sync"
)

type Evaluator struct {
	env      *cel.Env
	cache    sync.Map
	initOnce sync.Once
}

func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) init() error {
	var initErr error
	e.initOnce.Do(func() {
		env, err := cel.NewEnv(
			cel.Declarations(
				decls.NewVar("data", decls.NewMapType(decls.String, decls.Any)),
				decls.NewVar("metadata", decls.NewMapType(decls.String, decls.Any)),
			),
		)
		if err != nil {
			initErr = fmt.Errorf("initializing CEL environment: %w", err)
			return
		}
		e.env = env
	})
	return initErr
}

func (e *Evaluator) Evaluate(expr string, ctx *smctx.Context) (bool, error) {
	if err := e.init(); err != nil {
		return false, err
	}

	var prog cel.Program
	if cached, ok := e.cache.Load(expr); ok {
		prog = cached.(cel.Program)
	} else {
		ast, iss := e.env.Parse(expr)
		if iss.Err() != nil {
			return false, fmt.Errorf("failed to parse expression: %w", iss.Err())
		}

		checked, iss := e.env.Check(ast)
		if iss.Err() != nil {
			return false, fmt.Errorf("failed to check expression: %w", iss.Err())
		}

		prg, err := e.env.Program(checked)
		if err != nil {
			return false, fmt.Errorf("failed to create program: %w", err)
		}

		e.cache.Store(expr, prg)
		prog = prg
	}

	vars := map[string]interface{}{
		"data":     ctx.GetAllData(),
		"metadata": map[string]interface{}{},
	}

	val, _, err := prog.Eval(vars)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	result, ok := val.Value().(bool)
	if !ok {
		return false, fmt.Errorf("expression must evaluate to boolean")
	}

	return result, nil
}
