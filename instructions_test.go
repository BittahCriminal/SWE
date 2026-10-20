package swe

import (
	"os"
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

func TestQualitySkillsCoverCorrectnessBeforeLeanness(t *testing.T) {
	for path, wants := range map[string][]string{
		"skills/swe-review/SKILL.md": {"**Must fix:**", "If we skip it", "expected load"},
		"skills/swe-audit/SKILL.md":  {"What this repo does:", "security holes", "at most 20"},
		"skills/swe-debt/SKILL.md":   {"git grep --no-index --exclude-standard", "shortcut:"},
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range wants {
			if !strings.Contains(string(data), want) {
				t.Errorf("%s missing %q", path, want)
			}
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
	for _, want := range []string{"## Before you write", "List every place your change must reach", "shortcut: <the limit>"} {
		if !strings.Contains(full, want) {
			t.Fatalf("full instructions missing %q", want)
		}
	}
	if strings.Contains(full, "| **lite**") || strings.Contains(full, "| **ultra**") {
		t.Fatal("full instructions include another level")
	}
}
