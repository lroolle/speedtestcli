package cmdutil

import (
	"context"
	"fmt"
	"testing"
)

func TestExitCode_Nil(t *testing.T) {
	if got := ExitCode(nil); got != ExitSuccess {
		t.Errorf("expected %d, got %d", ExitSuccess, got)
	}
}

func TestExitCode_DeadlineExceeded(t *testing.T) {
	if got := ExitCode(context.DeadlineExceeded); got != ExitTimeout {
		t.Errorf("expected %d, got %d", ExitTimeout, got)
	}
}

func TestExitCode_WrappedDeadline(t *testing.T) {
	err := fmt.Errorf("test failed: %w", context.DeadlineExceeded)
	if got := ExitCode(err); got != ExitTimeout {
		t.Errorf("expected %d, got %d", ExitTimeout, got)
	}
}

func TestExitCode_ArgsError(t *testing.T) {
	err := &ErrArgs{Msg: "invalid backend"}
	if got := ExitCode(err); got != ExitArgs {
		t.Errorf("expected %d, got %d", ExitArgs, got)
	}
}

func TestExitCode_WrappedArgsError(t *testing.T) {
	err := fmt.Errorf("config: %w", &ErrArgs{Msg: "bad value"})
	if got := ExitCode(err); got != ExitArgs {
		t.Errorf("expected %d, got %d", ExitArgs, got)
	}
}

func TestExitCode_NetworkError(t *testing.T) {
	err := fmt.Errorf("connection refused")
	if got := ExitCode(err); got != ExitNetwork {
		t.Errorf("expected %d, got %d", ExitNetwork, got)
	}
}

func TestErrArgs_Error(t *testing.T) {
	err := &ErrArgs{Msg: "test"}
	if err.Error() != "test" {
		t.Errorf("expected 'test', got %q", err.Error())
	}
}

func TestErrNetwork_Error(t *testing.T) {
	err := &ErrNetwork{Msg: "timeout"}
	if err.Error() != "timeout" {
		t.Errorf("expected 'timeout', got %q", err.Error())
	}
}
