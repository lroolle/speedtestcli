package ooklabackend

import "testing"

func TestBackend_Name(t *testing.T) {
	b := New()
	if b.Name() != "ookla" {
		t.Errorf("expected 'ookla', got %q", b.Name())
	}
}
