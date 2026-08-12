package swe

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"
)

const maxHookInput = 1 << 20

type hookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

type hookOutput struct {
	SystemMessage      string              `json:"systemMessage,omitempty"`
	AdditionalContext  string              `json:"additionalContext,omitempty"`
	HookSpecificOutput *hookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

// FormatHookOutput serializes context in the protocol expected by the active host.
func FormatHookOutput(event string, mode Mode, context string) []byte {
	host := DetectHost()
	var output hookOutput
	switch host {
	case HostCopilot:
		if event == "SessionStart" && context != "" {
			output.AdditionalContext = context
		}
	case HostCodex:
		output.SystemMessage = "SWE:" + strings.ToUpper(string(mode))
		if context != "" {
			output.HookSpecificOutput = &hookSpecificOutput{event, context}
		}
	case HostQoder:
		if context != "" {
			output.HookSpecificOutput = &hookSpecificOutput{event, context}
		}
	default:
		if event != "SubagentStart" {
			return []byte(context)
		}
		output.HookSpecificOutput = &hookSpecificOutput{event, context}
	}
	data, _ := json.Marshal(output)
	return data
}

// HandleSessionStart activates the configured default and emits its ruleset.
// State writes are best-effort so a lifecycle hook can never block the host.
func HandleSessionStart() []byte {
	mode := ResolveDefaultMode()
	if mode == ModeOff {
		_ = ClearMode()
		context := ""
		if DetectHost() == HostClaude {
			context = "OK"
		}
		return FormatHookOutput("SessionStart", mode, context)
	}
	_ = SetMode(mode)
	return FormatHookOutput("SessionStart", mode, BuildInstructions(mode))
}

type promptPayload struct {
	Prompt string `json:"prompt"`
}

type agentPayload struct {
	AgentType string `json:"agent_type"`
}

func decodeHookInput(reader io.Reader, value any) bool {
	if reader == nil {
		return false
	}
	data, err := io.ReadAll(io.LimitReader(reader, maxHookInput+1))
	if err != nil || len(data) > maxHookInput {
		return false
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	return json.Unmarshal(data, value) == nil
}

type promptAction int

const (
	actionNone promptAction = iota
	actionReport
	actionSet
	actionSetDefault
	actionInvalid
)

type parsedPrompt struct {
	action promptAction
	mode   Mode
}

func parsePromptCommand(prompt string) parsedPrompt {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(prompt)))
	if len(fields) == 0 {
		return parsedPrompt{}
	}
	head := fields[0]
	if strings.HasPrefix(head, "@") || strings.HasPrefix(head, "$") {
		head = "/" + head[1:]
	}
	if head == "/swe-review" || head == "/swe:swe-review" {
		return parsedPrompt{action: actionSet, mode: ModeReview}
	}
	if head != "/swe" && head != "/swe:swe" {
		return parsedPrompt{}
	}
	if len(fields) == 1 {
		return parsedPrompt{action: actionReport}
	}
	if fields[1] == "default" {
		if len(fields) == 3 {
			if mode, ok := ParseDefaultMode(fields[2]); ok {
				return parsedPrompt{action: actionSetDefault, mode: mode}
			}
		}
		return parsedPrompt{action: actionInvalid}
	}
	if len(fields) == 2 {
		if mode, ok := ParseDefaultMode(fields[1]); ok {
			return parsedPrompt{action: actionSet, mode: mode}
		}
	}
	return parsedPrompt{action: actionInvalid}
}

func activeOrDefault() Mode {
	if mode, ok := ReadMode(); ok {
		return mode
	}
	return ResolveDefaultMode()
}

// HandlePromptSubmit tracks slash-command mode changes and handles Qoder's
// per-turn injection behavior.
func HandlePromptSubmit(reader io.Reader) []byte {
	var payload promptPayload
	if !decodeHookInput(reader, &payload) {
		return nil
	}
	action := parsePromptCommand(payload.Prompt)
	host := DetectHost()
	mode := activeOrDefault()
	context := ""
	changed := false
	deactivated := false

	switch action.action {
	case actionReport:
		context = "SWE MODE ACTIVE — level: " + string(mode)
	case actionSetDefault:
		if SaveDefaultMode(action.mode) == nil {
			mode = action.mode
			context = "SWE DEFAULT SET — new sessions start in " + string(mode) + "."
		}
	case actionSet:
		mode = action.mode
		if mode == ModeOff {
			_ = ClearMode()
			deactivated = true
			context = "SWE MODE OFF"
		} else {
			_ = SetMode(mode)
			changed = true
			context = "SWE MODE CHANGED — level: " + string(mode)
		}
	case actionInvalid:
		context = "Usage: /swe [lite|full|ultra|off] or /swe default <mode>"
	default:
		if IsDeactivationCommand(payload.Prompt) {
			mode = ModeOff
			_ = ClearMode()
			deactivated = true
			context = "SWE MODE OFF"
		}
	}

	if host == HostQoder && !deactivated && action.action != actionSetDefault && action.action != actionInvalid {
		mode = activeOrDefault()
		if mode != ModeOff {
			_ = SetMode(mode)
			prefix := ""
			if changed {
				prefix = "SWE MODE CHANGED — level: " + string(mode) + "\n\n"
			} else if action.action == actionReport {
				prefix = "SWE MODE ACTIVE — level: " + string(mode) + "\n\n"
			}
			context = prefix + BuildInstructions(mode)
		}
	}
	if context == "" {
		return nil
	}
	return FormatHookOutput("UserPromptSubmit", mode, context)
}

// HandleSubagentStart injects the active ruleset, optionally scoped by a
// case-insensitive SWE_SUBAGENT_MATCHER regular expression. Invalid or missing
// metadata fails open so the parent mode is not silently lost.
func HandleSubagentStart(reader io.Reader) []byte {
	mode, ok := ReadMode()
	if !ok || mode == ModeOff {
		return nil
	}
	if pattern := os.Getenv("SWE_SUBAGENT_MATCHER"); pattern != "" {
		matcher, err := regexp.Compile("(?i)" + pattern)
		if err == nil {
			var payload agentPayload
			if decodeHookInput(reader, &payload) && strings.TrimSpace(payload.AgentType) != "" && !matcher.MatchString(payload.AgentType) {
				return nil
			}
		}
	}
	return FormatHookOutput("SubagentStart", mode, BuildInstructions(mode))
}

// Badge returns the compact active-mode indicator used by status lines.
func Badge() string {
	mode, ok := ReadMode()
	if !ok || mode == ModeOff {
		return ""
	}
	if mode == ModeFull {
		return "[SWE]"
	}
	return "[SWE:" + strings.ToUpper(string(mode)) + "]"
}
