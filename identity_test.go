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
		".plugin/plugin.json": {
			`"name": "swe"`,
			`"hooks": "./hooks/copilot-hooks.json"`,
			`"repository": "https://github.com/bittahcriminal/SWE"`,
		},
		"gemini-extension.json": {
			`"name": "swe"`,
			`"description": "Smallest-correct-change engineering guidance, powered by Go."`,
		},
		"README.md": {
			"# SWE",
			"go install github.com/bittahcriminal/SWE/cmd/swe@latest",
			"copilot plugin marketplace add bittahcriminal/SWE",
			"copilot plugin install swe@swe",
			"gemini extensions install https://github.com/bittahcriminal/SWE",
			"swe install opencode",
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
