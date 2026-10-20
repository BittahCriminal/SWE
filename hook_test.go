package swe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func decodeOutput(t *testing.T, data []byte) hookOutput {
	t.Helper()
	var output hookOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("invalid hook JSON %q: %v", data, err)
	}
	return output
}

func TestCodexLifecycleAndPromptCommands(t *testing.T) {
	home := isolateEnvironment(t)
	pluginData := filepath.Join(home, "plugin-data")
	t.Setenv("PLUGIN_DATA", pluginData)
	t.Setenv("SWE_DEFAULT_MODE", "ultra")

	output := decodeOutput(t, HandleSessionStart())
	if output.SystemMessage != "" || output.HookSpecificOutput == nil || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "level: ultra") {
		t.Fatalf("unexpected session output: %#v", output)
	}
	if output.AdditionalContext != "" {
		t.Fatal("Codex context must not be emitted at the top level")
	}

	output = decodeOutput(t, HandlePromptSubmit(strings.NewReader(`{"prompt":"@swe lite"}`)))
	if output.SystemMessage != "" || output.HookSpecificOutput == nil || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "## Before you write") {
		t.Fatalf("mode switch output: %#v", output)
	}
	if mode, ok := ReadMode(); !ok || mode != ModeLite {
		t.Fatalf("active mode = %q, %v; want lite", mode, ok)
	}

	output = decodeOutput(t, HandlePromptSubmit(strings.NewReader(`{"prompt":"@swe"}`)))
	if output.HookSpecificOutput == nil || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "level: lite") {
		t.Fatalf("bare @swe reset or failed to report: %#v", output)
	}

	if got := HandlePromptSubmit(strings.NewReader(`{"prompt":"add a normal mode toggle"}`)); len(got) != 0 {
		t.Fatalf("ordinary prompt emitted hook output: %q", got)
	}
	if mode, ok := ReadMode(); !ok || mode != ModeLite {
		t.Fatal("incidental phrase deactivated SWE")
	}

	output = decodeOutput(t, HandlePromptSubmit(strings.NewReader(`{"prompt":"normal mode"}`)))
	if output.SystemMessage != "" || output.HookSpecificOutput == nil || output.HookSpecificOutput.AdditionalContext != "SWE MODE OFF" {
		t.Fatalf("deactivation output: %#v", output)
	}
	if _, ok := ReadMode(); ok {
		t.Fatal("deactivation left state behind")
	}
}

func TestBareCommandActivatesAndReviewStaysOneShot(t *testing.T) {
	isolateEnvironment(t)
	t.Setenv("PLUGIN_DATA", t.TempDir())
	t.Setenv("SWE_DEFAULT_MODE", "off")
	output := decodeOutput(t, HandlePromptSubmit(strings.NewReader(`{"prompt":"/swe"}`)))
	if output.HookSpecificOutput == nil || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "level: full") {
		t.Fatalf("bare /swe did not activate full: %#v", output)
	}
	if mode, ok := ReadMode(); !ok || mode != ModeFull {
		t.Fatalf("active mode = %q, %v; want full", mode, ok)
	}
	if got := HandlePromptSubmit(strings.NewReader(`{"prompt":"/swe-review"}`)); len(got) != 0 {
		t.Fatalf("review skill was intercepted: %q", got)
	}
	if mode, ok := ReadMode(); !ok || mode != ModeFull {
		t.Fatalf("review changed session mode to %q, %v", mode, ok)
	}
}

func TestCompactionKeepsLiveModeAndOff(t *testing.T) {
	isolateEnvironment(t)
	t.Setenv("PLUGIN_DATA", t.TempDir())
	t.Setenv("SWE_DEFAULT_MODE", "full")
	if err := SetMode(ModeUltra); err != nil {
		t.Fatal(err)
	}
	output := decodeOutput(t, HandleSessionStartInput(strings.NewReader(`{"source":"compact"}`)))
	if output.HookSpecificOutput == nil || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "level: ultra") {
		t.Fatalf("compaction reset live mode: %#v", output)
	}
	if err := ClearMode(); err != nil {
		t.Fatal(err)
	}
	if got := HandleSessionStartInput(strings.NewReader(`{"source":"compact"}`)); len(got) != 0 {
		t.Fatalf("compaction reactivated an off session: %q", got)
	}
}

func TestOffSessionStartIsSilent(t *testing.T) {
	isolateEnvironment(t)
	t.Setenv("SWE_DEFAULT_MODE", "off")
	if got := HandleSessionStart(); len(got) != 0 {
		t.Fatalf("off startup output = %q", got)
	}
}

func TestSessionStartIncludesBoundedCodebaseMap(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "internal", "reuse.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("package internal\nfunc ExistingHelper() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	isolateEnvironment(t)
	t.Setenv("PLUGIN_DATA", t.TempDir())
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv("SWE_MAP", "1")
	output := decodeOutput(t, HandleSessionStart())
	if output.HookSpecificOutput == nil || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "internal/: ExistingHelper") {
		t.Fatalf("session map missing: %#v", output)
	}
}

func TestDefaultCommandDoesNotChangeSession(t *testing.T) {
	isolateEnvironment(t)
	if err := SetMode(ModeUltra); err != nil {
		t.Fatal(err)
	}
	output := HandlePromptSubmit(strings.NewReader(`{"prompt":"/swe default lite"}`))
	if !strings.Contains(string(output), "SWE DEFAULT SET") {
		t.Fatalf("missing confirmation: %q", output)
	}
	if mode, ok := ReadMode(); !ok || mode != ModeUltra {
		t.Fatalf("default command changed session to %q", mode)
	}
	if got := ResolveDefaultMode(); got != ModeLite {
		t.Fatalf("saved default = %q, want lite", got)
	}
}

func TestNativeClaudeAndCopilotOutput(t *testing.T) {
	home := isolateEnvironment(t)
	t.Setenv("SWE_DEFAULT_MODE", "full")
	claude := HandleSessionStart()
	if !strings.HasPrefix(string(claude), "SWE MODE ACTIVE — level: full") {
		t.Fatalf("Claude should receive raw context, got %q", claude)
	}
	subagent := decodeOutput(t, HandleSubagentStart(nil))
	if subagent.HookSpecificOutput == nil || subagent.HookSpecificOutput.HookEventName != "SubagentStart" {
		t.Fatalf("Claude subagent output: %#v", subagent)
	}

	t.Setenv("COPILOT_PLUGIN_DATA", filepath.Join(home, "copilot"))
	copilot := decodeOutput(t, HandleSessionStart())
	if !strings.Contains(copilot.AdditionalContext, "level: full") || copilot.SystemMessage != "" {
		t.Fatalf("Copilot session output: %#v", copilot)
	}
	prompt := decodeOutput(t, HandlePromptSubmit(strings.NewReader(`{"prompt":"/swe ultra"}`)))
	if prompt != (hookOutput{}) {
		t.Fatalf("Copilot prompt output must be empty JSON, got %#v", prompt)
	}
	if mode, ok := ReadMode(); !ok || mode != ModeUltra {
		t.Fatal("Copilot prompt did not update state")
	}
}

func TestGeminiSessionOutput(t *testing.T) {
	home := isolateEnvironment(t)
	t.Setenv("GEMINI_SESSION_ID", "gemini-test-session")
	t.Setenv("SWE_DEFAULT_MODE", "ultra")

	output := decodeOutput(t, HandleSessionStart())
	if output.SystemMessage != "SWE:ULTRA" || output.HookSpecificOutput == nil {
		t.Fatalf("unexpected Gemini session output: %#v", output)
	}
	if output.HookSpecificOutput.HookEventName != "SessionStart" || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "level: ultra") {
		t.Fatalf("Gemini context missing: %#v", output.HookSpecificOutput)
	}
	wantState := filepath.Join(home, ".gemini", stateFile)
	if data, err := os.ReadFile(wantState); err != nil || string(data) != "ultra" {
		t.Fatalf("Gemini state = %q, %v", data, err)
	}
}

func TestSubagentMatcher(t *testing.T) {
	isolateEnvironment(t)
	if err := SetMode(ModeFull); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SWE_SUBAGENT_MATCHER", "general|plan")
	if got := HandleSubagentStart(strings.NewReader(`{"agent_type":"Explore"}`)); len(got) != 0 {
		t.Fatalf("non-matching agent received context: %q", got)
	}
	if got := HandleSubagentStart(strings.NewReader(`{"agent_type":"General-purpose"}`)); len(got) == 0 {
		t.Fatal("matching agent did not receive context")
	}
	if got := HandleSubagentStart(strings.NewReader(`{}`)); len(got) == 0 {
		t.Fatal("missing agent_type must fail open")
	}
	t.Setenv("SWE_SUBAGENT_MATCHER", "(")
	if got := HandleSubagentStart(strings.NewReader(`{"agent_type":"anything"}`)); len(got) == 0 {
		t.Fatal("invalid matcher must fail open")
	}
}

func TestQoderInjectsEveryTurn(t *testing.T) {
	home := isolateEnvironment(t)
	t.Setenv("QODER_SESSION_ID", "test-session")
	t.Setenv("SWE_DEFAULT_MODE", "full")
	output := decodeOutput(t, HandlePromptSubmit(strings.NewReader(`{"prompt":"write a function"}`)))
	if output.HookSpecificOutput == nil || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "level: full") {
		t.Fatalf("Qoder did not inject rules: %#v", output)
	}
	if data, err := os.ReadFile(filepath.Join(home, ".qoder", stateFile)); err != nil || string(data) != "full" {
		t.Fatalf("Qoder state = %q, %v", data, err)
	}
	output = decodeOutput(t, HandlePromptSubmit(strings.NewReader(`{"prompt":"/swe ultra"}`)))
	if !strings.Contains(output.HookSpecificOutput.AdditionalContext, "MODE CHANGED") || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "level: ultra") {
		t.Fatalf("Qoder mode switch output: %#v", output)
	}
}

func TestQoderOffPersistsAcrossPrompts(t *testing.T) {
	isolateEnvironment(t)
	t.Setenv("QODER_SESSION_ID", "test-session")
	t.Setenv("QODER_CONFIG_DIR", t.TempDir())
	if got := HandlePromptSubmit(strings.NewReader(`{"prompt":"/swe off"}`)); len(got) == 0 {
		t.Fatal("off command produced no confirmation")
	}
	if got := HandlePromptSubmit(strings.NewReader(`{"prompt":"write a function"}`)); len(got) != 0 {
		t.Fatalf("Qoder reactivated after off: %q", got)
	}
}

func TestBadge(t *testing.T) {
	isolateEnvironment(t)
	if Badge() != "" {
		t.Fatal("inactive badge should be empty")
	}
	if err := SetMode(ModeFull); err != nil {
		t.Fatal(err)
	}
	if got := Badge(); got != "[SWE]" {
		t.Fatalf("full badge = %q", got)
	}
	if err := SetMode(ModeUltra); err != nil {
		t.Fatal(err)
	}
	if got := Badge(); got != "[SWE:ULTRA]" {
		t.Fatalf("ultra badge = %q", got)
	}
}

func TestOversizedHookInputIsIgnored(t *testing.T) {
	isolateEnvironment(t)
	input := strings.NewReader(strings.Repeat("x", maxHookInput+1))
	if got := HandlePromptSubmit(input); len(got) != 0 {
		t.Fatalf("oversized input produced output: %q", got)
	}
}
