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
	if output.SystemMessage != "SWE:ULTRA" || output.HookSpecificOutput == nil || !strings.Contains(output.HookSpecificOutput.AdditionalContext, "level: ultra") {
		t.Fatalf("unexpected session output: %#v", output)
	}
	if output.AdditionalContext != "" {
		t.Fatal("Codex context must not be emitted at the top level")
	}

	output = decodeOutput(t, HandlePromptSubmit(strings.NewReader(`{"prompt":"@swe lite"}`)))
	if output.SystemMessage != "SWE:LITE" {
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
	if output.SystemMessage != "SWE:OFF" {
		t.Fatalf("deactivation output: %#v", output)
	}
	if _, ok := ReadMode(); ok {
		t.Fatal("deactivation left state behind")
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
