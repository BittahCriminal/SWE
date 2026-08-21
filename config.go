package swe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const stateFile = ".swe-active"

// Host identifies the lifecycle-hook protocol expected by the caller.
type Host string

const (
	HostClaude  Host = "claude"
	HostCodex   Host = "codex"
	HostCopilot Host = "copilot"
	HostGemini  Host = "gemini"
	HostQoder   Host = "qoder"
)

// DetectHost follows the environment conventions used by the supported agents.
func DetectHost() Host {
	if os.Getenv("COPILOT_PLUGIN_DATA") != "" || isVSCodeCopilotRoot(os.Getenv("CLAUDE_PLUGIN_ROOT")) {
		return HostCopilot
	}
	if os.Getenv("PLUGIN_DATA") != "" {
		return HostCodex
	}
	if os.Getenv("GEMINI_SESSION_ID") != "" {
		return HostGemini
	}
	if os.Getenv("QODER_SESSION_ID") != "" {
		return HostQoder
	}
	return HostClaude
}

func isVSCodeCopilotRoot(root string) bool {
	root = "/" + strings.Trim(strings.ToLower(filepath.ToSlash(root)), "/") + "/"
	return strings.Contains(root, "/.vscode/") && strings.Contains(root, "/agent-plugins/")
}

func userHomeDir() (string, error) {
	if runtime.GOOS == "windows" {
		if home := os.Getenv("USERPROFILE"); home != "" {
			return home, nil
		}
	}
	return os.UserHomeDir()
}

// ConfigDir returns SWE's platform-appropriate persistent configuration directory.
func ConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "swe"), nil
	}
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		if dir := os.Getenv("APPDATA"); dir != "" {
			return filepath.Join(dir, "swe"), nil
		}
		return filepath.Join(home, "AppData", "Roaming", "swe"), nil
	}
	return filepath.Join(home, ".config", "swe"), nil
}

// ConfigPath returns the JSON configuration file path.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func claudeDir() (string, error) {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

// StatePath returns the host-scoped session-mode file.
func StatePath() (string, error) {
	var dir string
	var err error
	switch DetectHost() {
	case HostCodex:
		dir = os.Getenv("PLUGIN_DATA")
	case HostCopilot:
		dir = os.Getenv("COPILOT_PLUGIN_DATA")
		if dir == "" {
			dir, err = claudeDir()
		}
	case HostGemini:
		var home string
		home, err = userHomeDir()
		dir = filepath.Join(home, ".gemini")
	case HostQoder:
		var home string
		home, err = userHomeDir()
		dir = filepath.Join(home, ".qoder")
	default:
		dir, err = claudeDir()
	}
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", errors.New("could not resolve SWE state directory")
	}
	return filepath.Join(dir, stateFile), nil
}

// ResolveDefaultMode applies environment, config-file, then built-in precedence.
func ResolveDefaultMode() Mode {
	if mode, ok := ParseDefaultMode(os.Getenv("SWE_DEFAULT_MODE")); ok {
		return mode
	}
	path, err := ConfigPath()
	if err != nil {
		return DefaultMode
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultMode
	}
	var config map[string]any
	if json.Unmarshal(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}), &config) != nil {
		return DefaultMode
	}
	value, _ := config["defaultMode"].(string)
	if mode, ok := ParseDefaultMode(value); ok {
		return mode
	}
	return DefaultMode
}

// SaveDefaultMode merges a runtime default into config.json without discarding
// unrelated user settings.
func SaveDefaultMode(mode Mode) error {
	if parsed, ok := ParseDefaultMode(string(mode)); !ok || parsed != mode {
		return fmt.Errorf("invalid default mode %q", mode)
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	config := map[string]any{}
	if data, readErr := os.ReadFile(path); readErr == nil {
		_ = json.Unmarshal(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}), &config)
		if config == nil {
			config = map[string]any{}
		}
	}
	config["defaultMode"] = mode
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, append(data, '\n'), 0o600)
}

// ReadMode returns the active host-scoped mode. Missing or invalid state is inactive.
func ReadMode() (Mode, bool) {
	path, err := StatePath()
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	mode, ok := ParseMode(string(data))
	return mode, ok && mode != ModeOff
}

// SetMode persists a host-scoped session mode. Off removes the active state.
func SetMode(mode Mode) error {
	parsed, ok := ParseMode(string(mode))
	if !ok || parsed != mode {
		return fmt.Errorf("invalid mode %q", mode)
	}
	if mode == ModeOff {
		return ClearMode()
	}
	path, err := StatePath()
	if err != nil {
		return err
	}
	return writeAtomic(path, []byte(mode), 0o600)
}

// ClearMode removes the active host-scoped mode.
func ClearMode() error {
	path, err := StatePath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func writeAtomic(path string, data []byte, mode os.FileMode) (err error) {
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".swe-*")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer func() {
		_ = file.Close()
		_ = os.Remove(temp)
	}()
	if err = file.Chmod(mode); err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temp, path)
}
