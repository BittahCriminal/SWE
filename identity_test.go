package swe

import (
	"os"
	"strings"
	"testing"
)

func TestRepositoryUsesSWEIdentity(t *testing.T) {
	wants := map[string][]string{
		"go.mod": {
			"module github.com/bittahcriminal/SWE",
		},
		".codex-plugin/plugin.json": {
			`"name": "swe"`,
			`"displayName": "SWE"`,
			`"repository": "https://github.com/bittahcriminal/SWE"`,
		},
		".claude-plugin/plugin.json": {
			`"name": "swe"`,
			`"repository": "https://github.com/bittahcriminal/SWE"`,
		},
		"README.md": {
			"# SWE",
			"go install github.com/bittahcriminal/SWE/cmd/swe@latest",
		},
	}

	for path, expected := range wants {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		for _, want := range expected {
			if !strings.Contains(string(contents), want) {
				t.Errorf("%s does not contain %q", path, want)
			}
		}
	}
}
