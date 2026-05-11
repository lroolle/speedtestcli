package version

import "testing"

func TestFull_DevDefault(t *testing.T) {
	Version = "dev"
	GitCommit = "none"
	got := Full()
	if got != "dev" {
		t.Errorf("expected 'dev', got %q", got)
	}
}

func TestFull_WithShortCommit(t *testing.T) {
	Version = "v1.0.0"
	GitCommit = "abc1234"
	got := Full()
	if got != "v1.0.0 (abc1234)" {
		t.Errorf("expected 'v1.0.0 (abc1234)', got %q", got)
	}
}

func TestFull_TruncatesLongCommit(t *testing.T) {
	Version = "v1.0.0"
	GitCommit = "abc1234567890def"
	got := Full()
	if got != "v1.0.0 (abc1234)" {
		t.Errorf("expected 'v1.0.0 (abc1234)', got %q", got)
	}
}

func TestFull_EmptyCommit(t *testing.T) {
	Version = "v1.0.0"
	GitCommit = ""
	got := Full()
	if got != "v1.0.0" {
		t.Errorf("expected 'v1.0.0', got %q", got)
	}
}
