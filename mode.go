package swe

import "strings"

// Mode controls how aggressively SWE favors the smallest correct solution.
type Mode string

const (
	ModeOff    Mode = "off"
	ModeLite   Mode = "lite"
	ModeFull   Mode = "full"
	ModeUltra  Mode = "ultra"
	ModeReview Mode = "review"

	DefaultMode Mode = ModeFull
)

// ParseMode accepts every session mode, including the review-only mode.
func ParseMode(value string) (Mode, bool) {
	mode := Mode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case ModeOff, ModeLite, ModeFull, ModeUltra, ModeReview:
		return mode, true
	default:
		return "", false
	}
}

// ParseRuntimeMode accepts modes that may persist for a session. Review is a
// one-shot skill, not a session level.
func ParseRuntimeMode(value string) (Mode, bool) {
	mode, ok := ParseMode(value)
	return mode, ok && mode != ModeReview
}

// ParseDefaultMode accepts modes that may be persisted as a startup default.
// Review is intentionally session-only.
func ParseDefaultMode(value string) (Mode, bool) {
	return ParseRuntimeMode(value)
}

// IsDeactivationCommand matches only standalone deactivation commands.
func IsDeactivationCommand(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimRight(value, ".!? \t\r\n")
	return value == "stop swe" || value == "normal mode"
}
