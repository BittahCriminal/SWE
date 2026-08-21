package swe

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	openCodeBlockStart = "<!-- SWE START -->"
	openCodeBlockEnd   = "<!-- SWE END -->"
)

//go:embed AGENTS.md skills/*/SKILL.md integrations/opencode/commands/*.md
var openCodeAssets embed.FS

// OpenCodeConfigDir returns the global OpenCode configuration directory.
func OpenCodeConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "opencode"), nil
	}
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "opencode"), nil
}

// InstallOpenCode installs SWE rules, skills, and commands without replacing
// unrelated OpenCode configuration.
func InstallOpenCode(configDir string) error {
	if strings.TrimSpace(configDir) == "" {
		return errors.New("OpenCode config directory is empty")
	}
	if err := installOpenCodeRules(configDir); err != nil {
		return err
	}
	if err := copyEmbeddedFiles(configDir, "skills/*/SKILL.md", ""); err != nil {
		return err
	}
	return copyEmbeddedFiles(configDir, "integrations/opencode/commands/*.md", "commands")
}

func installOpenCodeRules(configDir string) error {
	rules, err := openCodeAssets.ReadFile("AGENTS.md")
	if err != nil {
		return fmt.Errorf("read embedded SWE rules: %w", err)
	}
	path := filepath.Join(configDir, "AGENTS.md")
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read OpenCode rules: %w", err)
	}
	merged, err := mergeSWEBlock(string(existing), string(rules))
	if err != nil {
		return err
	}
	if err := writeAtomic(path, []byte(merged), 0o600); err != nil {
		return fmt.Errorf("write OpenCode rules: %w", err)
	}
	return nil
}

func mergeSWEBlock(existing, rules string) (string, error) {
	start := strings.Index(existing, openCodeBlockStart)
	end := strings.Index(existing, openCodeBlockEnd)
	if (start == -1) != (end == -1) || (start >= 0 && end < start) {
		return "", errors.New("OpenCode AGENTS.md contains an incomplete SWE block")
	}
	block := openCodeBlockStart + "\n" + strings.TrimSpace(rules) + "\n" + openCodeBlockEnd
	if start >= 0 {
		end += len(openCodeBlockEnd)
		before := strings.TrimRight(existing[:start], "\r\n")
		after := strings.TrimLeft(existing[end:], "\r\n")
		if before != "" {
			before += "\n\n"
		}
		if after != "" {
			after = "\n\n" + after
		}
		return before + block + after + "\n", nil
	}
	prefix := strings.TrimRight(existing, "\r\n")
	if prefix != "" {
		prefix += "\n\n"
	}
	return prefix + block + "\n", nil
}

func copyEmbeddedFiles(configDir, pattern, destinationRoot string) error {
	paths, err := fs.Glob(openCodeAssets, pattern)
	if err != nil {
		return fmt.Errorf("find embedded OpenCode assets: %w", err)
	}
	for _, source := range paths {
		data, err := openCodeAssets.ReadFile(source)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", source, err)
		}
		relative := source
		if destinationRoot != "" {
			relative = filepath.Join(destinationRoot, filepath.Base(source))
		}
		destination := filepath.Join(configDir, filepath.FromSlash(relative))
		if err := writeAtomic(destination, data, 0o600); err != nil {
			return fmt.Errorf("write %s: %w", destination, err)
		}
	}
	return nil
}
