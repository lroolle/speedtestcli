package cmdutil

import (
	"context"
	"errors"
)

const (
	ExitSuccess = 0
	ExitNetwork = 1
	ExitArgs    = 2
	ExitTimeout = 3
)

type ErrArgs struct{ Msg string }

func (e *ErrArgs) Error() string { return e.Msg }

type ErrNetwork struct{ Msg string }

func (e *ErrNetwork) Error() string { return e.Msg }

func ExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ExitTimeout
	}
	var argsErr *ErrArgs
	if errors.As(err, &argsErr) {
		return ExitArgs
	}
	return ExitNetwork
}
