package handler

import "testing"

func TestOptionalExcerptLeavesBlankInputBlank(t *testing.T) {
	if got := optionalExcerpt("   \n\t"); got != "" {
		t.Fatalf("optionalExcerpt() = %q, want empty string", got)
	}
}

func TestOptionalExcerptKeepsManualCopy(t *testing.T) {
	if got := optionalExcerpt("  手写摘要  "); got != "手写摘要" {
		t.Fatalf("optionalExcerpt() = %q, want trimmed manual excerpt", got)
	}
}
