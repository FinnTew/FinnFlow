package statemachine

import (
	"errors"
)

var (
	ErrStateNotFound        = errors.New("state not found")
	ErrInvalidTransition    = errors.New("invalid transition")
	ErrInvalidExpression    = errors.New("invalid expression")
	ErrContextRequired      = errors.New("ctx is required")
	ErrInitialStateRequired = errors.New("initial state is required")
	ErrDuplicateState       = errors.New("duplicate state")
	ErrInvalidConfig        = errors.New("invalid config")
)
