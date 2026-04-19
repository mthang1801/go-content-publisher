package buildinfo

import "testing"

func TestShortCommit(t *testing.T) {
	if got := shortCommit("1234567890abcdef"); got != "1234567890ab" {
		t.Fatalf("expected shortened commit, got %q", got)
	}
	if got := shortCommit("1234"); got != "1234" {
		t.Fatalf("expected unchanged short commit, got %q", got)
	}
}
