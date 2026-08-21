package swe

import (
	"encoding/json"
	"os"
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
		if version == "" {
			version = manifest.Version
		}
		if manifest.Version == "" || manifest.Version != version {
			t.Errorf("%s version = %q, want %q", path, manifest.Version, version)
		}
	}
}
