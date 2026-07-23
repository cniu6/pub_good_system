package models

import "testing"

func TestBuildCertUniqueKey(t *testing.T) {
	if got := BuildCertUniqueKey("abc123", 0, false); got == nil || *got != "ABC123" {
		t.Fatalf("pending should keep unique key, got %#v", got)
	}
	if got := BuildCertUniqueKey("abc123", 1, false); got == nil || *got != "ABC123" {
		t.Fatalf("approved should keep unique key, got %#v", got)
	}
	if got := BuildCertUniqueKey("abc123", 2, false); got != nil {
		t.Fatalf("rejected should clear unique key")
	}
	if got := BuildCertUniqueKey("abc123", 1, true); got != nil {
		t.Fatalf("deleted should clear unique key")
	}
}
