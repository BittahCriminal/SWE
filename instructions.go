package swe

import (
	_ "embed"
	"regexp"
	"strings"
)

//go:embed skills/swe/SKILL.md
var mainSkill string

//go:embed skills/swe-review/SKILL.md
var reviewSkill string

var (
	frontmatterPattern = regexp.MustCompile(`(?s)^---\r?\n.*?\r?\n---\s*`)
	tableModePattern   = regexp.MustCompile(`^\|\s*\*\*(.+?)\*\*\s*\|`)
	exampleModePattern = regexp.MustCompile(`^-\s*([^:]+):\s*"`)
)

func stripFrontmatter(body string) string {
	return frontmatterPattern.ReplaceAllString(body, "")
}

// FilterSkillBodyForMode retains the selected intensity row and example while
// leaving ordinary prose untouched.
func FilterSkillBodyForMode(body string, mode Mode) string {
	if mode != ModeLite && mode != ModeFull && mode != ModeUltra {
		mode = DefaultMode
	}
	lines := strings.Split(stripFrontmatter(body), "\n")
	kept := lines[:0]
	for _, line := range lines {
		if match := tableModePattern.FindStringSubmatch(line); match != nil {
			if rowMode, ok := ParseDefaultMode(match[1]); ok && rowMode != mode {
				continue
			}
		}
		if match := exampleModePattern.FindStringSubmatch(line); match != nil {
			if exampleMode, ok := ParseDefaultMode(match[1]); ok && exampleMode != mode {
				continue
			}
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// BuildInstructions returns the host-ready instruction context for a mode.
func BuildInstructions(mode Mode) string {
	switch mode {
	case ModeOff:
		return ""
	case ModeReview:
		return "SWE MODE ACTIVE — level: review\n\n" + stripFrontmatter(reviewSkill)
	case ModeLite, ModeFull, ModeUltra:
		return "SWE MODE ACTIVE — level: " + string(mode) + "\n\n" + FilterSkillBodyForMode(mainSkill, mode)
	default:
		return BuildInstructions(DefaultMode)
	}
}
