package swe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallOpenCodePreservesRulesAndIsIdempotent(t *testing.T) {
	configDir := t.TempDir()
	agentsPath := filepath.Join(configDir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("# Existing rules\n\nKeep this.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := InstallOpenCode(configDir); err != nil {
		t.Fatal(err)
	}
	if err := InstallOpenCode(configDir); err != nil {
		t.Fatalf("second install: %v", err)
	}

	agents, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(agents)
	for _, want := range []string{"# Existing rules", "Keep this.", "# SWE, lazy senior dev mode"} {
		if !strings.Contains(text, want) {
			t.Errorf("AGENTS.md missing %q:\n%s", want, text)
		}
	}
	if got := strings.Count(text, "<!-- SWE START -->"); got != 1 {
		t.Errorf("SWE block count = %d, want 1", got)
	}

	for _, path := range []string{
		filepath.Join(configDir, "skills", "swe", "SKILL.md"),
		filepath.Join(configDir, "skills", "swe-review", "SKILL.md"),
		filepath.Join(configDir, "skills", "swe-audit", "SKILL.md"),
		filepath.Join(configDir, "commands", "swe.md"),
		filepath.Join(configDir, "commands", "swe-review.md"),
		filepath.Join(configDir, "commands", "swe-audit.md"),
		filepath.Join(configDir, "commands", "swe-debt.md"),
		filepath.Join(configDir, "commands", "swe-gain.md"),
		filepath.Join(configDir, "commands", "swe-help.md"),
	} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read %s: %v", path, err)
			continue
		}
		if !strings.Contains(strings.ToLower(string(contents)), "swe") {
			t.Errorf("%s does not identify SWE", path)
		}
	}
}
