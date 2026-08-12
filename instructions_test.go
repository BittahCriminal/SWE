package swe

import (
	"strings"
	"testing"
)

func TestInstructionFiltering(t *testing.T) {
	body := `---
name: test
---
| **lite** | lite row |
| **full** | full row |
| **ultra** | ultra row |
- lite: "lite example"
- full: "full example"
- ultra: "ultra example"
- Full: ordinary prose must survive
`
	got := FilterSkillBodyForMode(body, ModeLite)
	for _, want := range []string{"lite row", "lite example", "ordinary prose must survive"} {
		if !strings.Contains(got, want) {
			t.Errorf("filtered body missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"full row", "ultra row", "full example", "ultra example"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("filtered body unexpectedly contains %q:\n%s", unwanted, got)
		}
	}
}

func TestBuildInstructions(t *testing.T) {
	for _, mode := range []Mode{ModeLite, ModeFull, ModeUltra, ModeReview} {
		got := BuildInstructions(mode)
		if !strings.HasPrefix(got, "SWE MODE ACTIVE — level: "+string(mode)) {
			t.Fatalf("%s header missing: %q", mode, got)
		}
	}
	if got := BuildInstructions(ModeOff); got != "" {
		t.Fatalf("off instructions = %q, want empty", got)
	}
	full := BuildInstructions(ModeFull)
	if !strings.Contains(full, "Used the existing cache helper") || strings.Contains(full, "Cache added") || strings.Contains(full, "No cache until profiling") {
		t.Fatalf("full instructions contain wrong examples")
	}
}
