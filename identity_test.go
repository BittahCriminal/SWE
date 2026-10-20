package swe

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// The four host manifests must agree on plugin name and version.
func TestManifestNameAndVersionSync(t *testing.T) {
	paths := []string{
		".claude-plugin/plugin.json",
		".codex-plugin/plugin.json",
		".plugin/plugin.json",
		"gemini-extension.json",
	}
	version := ""
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var manifest struct{ Name, Version string }
		if err := json.Unmarshal(data, &manifest); err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		if manifest.Name != "swe" {
			t.Errorf("%s name = %q, want %q", path, manifest.Name, "swe")
		}
		if manifest.Version != "0.3.0" {
			t.Errorf("%s version = %q, want 0.3.0", path, manifest.Version)
		}
		if version == "" {
			version = manifest.Version
		}
		if manifest.Version == "" || manifest.Version != version {
			t.Errorf("%s version = %q, want %q", path, manifest.Version, version)
		}
	}
}

func TestHostManifestPaths(t *testing.T) {
	for path, want := range map[string]string{
		".claude-plugin/plugin.json": `"hooks": "./hooks/claude-codex-hooks.json"`,
		".codex-plugin/plugin.json":  `"hooks": "./hooks/claude-codex-hooks.json"`,
		"gemini-extension.json":      `"contextFileName": "AGENTS.md"`,
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), want) {
			t.Errorf("%s missing %s", path, want)
		}
	}
	if _, err := os.Stat("hooks/hooks.json"); !os.IsNotExist(err) {
		t.Fatalf("Gemini-reserved hooks/hooks.json must not exist: %v", err)
	}
}
