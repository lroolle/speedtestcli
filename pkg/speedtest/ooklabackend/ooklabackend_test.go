package ooklabackend

import "testing"

func TestBackend_Name(t *testing.T) {
	b := New()
	if b.Name() != "ookla" {
		t.Errorf("expected 'ookla', got %q", b.Name())
	}
}

func TestBackend_Granularity(t *testing.T) {
	b := New()
	if b.Granularity() != "aggregate" {
		t.Errorf("expected 'aggregate', got %q", b.Granularity())
	}
}

func TestWithDirect_CreatesBackend(t *testing.T) {
	b := New(WithDirect())
	if b.Name() != "ookla" {
		t.Errorf("expected 'ookla', got %q", b.Name())
	}
	if b.client == nil {
		t.Error("client should not be nil after WithDirect")
	}
}

func TestNew_DefaultClient(t *testing.T) {
	b := New()
	if b.client == nil {
		t.Error("default client should not be nil")
	}
}
