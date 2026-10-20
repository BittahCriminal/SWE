package swe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func isolateEnvironment(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("PLUGIN_DATA", "")
	t.Setenv("COPILOT_PLUGIN_DATA", "")
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	t.Setenv("QODER_SESSION_ID", "")
	t.Setenv("QODER_CONFIG_DIR", "")
	t.Setenv("GEMINI_SESSION_ID", "")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Setenv("GEMINI_PROJECT_DIR", "")
	t.Setenv("SWE_DEFAULT_MODE", "")
	t.Setenv("SWE_SUBAGENT_MATCHER", "")
	t.Setenv("SWE_MAP", "0")
	return home
}

func TestDefaultModePrecedenceAndMerge(t *testing.T) {
	isolateEnvironment(t)
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append([]byte{0xef, 0xbb, 0xbf}, []byte(`{"defaultMode":"lite","customSetting":42}`)...), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := ResolveDefaultMode(); got != ModeLite {
		t.Fatalf("ResolveDefaultMode() = %q, want lite", got)
	}
	if err := SaveDefaultMode(ModeUltra); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	if config["defaultMode"] != "ultra" || config["customSetting"] != float64(42) {
		t.Fatalf("config merge lost data: %#v", config)
	}
	t.Setenv("SWE_DEFAULT_MODE", "full")
	if got := ResolveDefaultMode(); got != ModeFull {
		t.Fatalf("environment did not win: %q", got)
	}
	t.Setenv("SWE_DEFAULT_MODE", "review")
	if got := ResolveDefaultMode(); got != ModeUltra {
		t.Fatalf("review must not be a default: %q", got)
	}
	if err := SaveDefaultMode(ModeReview); err == nil {
		t.Fatal("SaveDefaultMode(review) unexpectedly succeeded")
	}
}

func TestStateIsHostScoped(t *testing.T) {
	home := isolateEnvironment(t)
	codex := filepath.Join(home, "codex-data")
	t.Setenv("PLUGIN_DATA", codex)
	if DetectHost() != HostCodex {
		t.Fatalf("DetectHost() = %q, want codex", DetectHost())
	}
	if err := SetMode(ModeUltra); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(codex, stateFile)); err != nil || string(data) != "ultra" {
		t.Fatalf("codex state = %q, %v", data, err)
	}

	copilot := filepath.Join(home, "copilot-data")
	t.Setenv("COPILOT_PLUGIN_DATA", copilot)
	if DetectHost() != HostCopilot {
		t.Fatalf("copilot must take precedence over codex, got %q", DetectHost())
	}
	if err := SetMode(ModeLite); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(copilot, stateFile)); err != nil || string(data) != "lite" {
		t.Fatalf("copilot state = %q, %v", data, err)
	}
	if err := SetMode(ModeOff); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(copilot, stateFile)); !os.IsNotExist(err) {
		t.Fatalf("off should remove state, got %v", err)
	}
}

func TestVSCodeCopilotFallback(t *testing.T) {
	home := isolateEnvironment(t)
	t.Setenv("CLAUDE_PLUGIN_ROOT", filepath.Join(home, ".vscode", "agent-plugins", "github.com", "bittahcriminal", "SWE"))
	if DetectHost() != HostCopilot {
		t.Fatalf("DetectHost() = %q, want copilot", DetectHost())
	}
	path, err := StatePath()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".claude", stateFile); path != want {
		t.Fatalf("StatePath() = %q, want %q", path, want)
	}
}

func TestDetectHostGemini(t *testing.T) {
	home := isolateEnvironment(t)
	t.Setenv("GEMINI_SESSION_ID", "gemini-test-session")
	if DetectHost() != HostGemini {
		t.Fatalf("DetectHost() = %q, want gemini", DetectHost())
	}
	want := filepath.Join(home, ".gemini", stateFile)
	if got, err := StatePath(); err != nil || got != want {
		t.Fatalf("StatePath() = %q, %v; want %q", got, err, want)
	}
}

func TestStateIsProjectScoped(t *testing.T) {
	home := isolateEnvironment(t)
	projectA := filepath.Join(home, "project-a")
	t.Setenv("CLAUDE_PROJECT_DIR", projectA)
	if err := SetMode(ModeLite); err != nil {
		t.Fatal(err)
	}
	projectState, err := StatePath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(projectState) != filepath.Join(home, ".claude", "swe-modes") {
		t.Fatalf("project state path = %q", projectState)
	}
	if data, err := os.ReadFile(projectState); err != nil || string(data) != "lite" {
		t.Fatalf("project state = %q, %v", data, err)
	}
	if data, err := os.ReadFile(filepath.Join(home, ".claude", stateFile)); err != nil || string(data) != "lite" {
		t.Fatalf("shared state = %q, %v", data, err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", filepath.Join(home, "project-b"))
	if mode, ok := ReadMode(); ok {
		t.Fatalf("project B inherited project A mode %q", mode)
	}
}

func TestQoderConfigDirAndOffState(t *testing.T) {
	isolateEnvironment(t)
	configDir := filepath.Join(t.TempDir(), "qoder-config")
	t.Setenv("QODER_SESSION_ID", "session")
	t.Setenv("QODER_CONFIG_DIR", configDir)
	if err := SetMode(ModeOff); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(configDir, stateFile))
	if err != nil || string(data) != "off" {
		t.Fatalf("Qoder off state = %q, %v", data, err)
	}
}
