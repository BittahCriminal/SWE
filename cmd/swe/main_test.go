package main

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunHelpAndValidation(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(nil, strings.NewReader(""), &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("help: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"instructions", "off"}, strings.NewReader(""), &stdout, &stderr); code != 2 {
		t.Fatalf("invalid instructions mode code=%d", code)
	}
}

func TestRunStatus(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("PLUGIN_DATA", "")
	t.Setenv("COPILOT_PLUGIN_DATA", "")
	t.Setenv("QODER_SESSION_ID", "")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"status"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("status: code=%d stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); got != "mode=off default=full\n" {
		t.Fatalf("status = %q", got)
	}
}

func TestInputHookNeverHangs(t *testing.T) {
	previous := hookInputTimeout
	hookInputTimeout = 10 * time.Millisecond
	t.Cleanup(func() { hookInputTimeout = previous })
	reader, writer := io.Pipe()
	t.Cleanup(func() {
		_ = reader.Close()
		_ = writer.Close()
	})
	var stdout, stderr bytes.Buffer
	started := time.Now()
	if code := run([]string{"hook", "prompt-submit"}, reader, &stdout, &stderr); code != 0 {
		t.Fatalf("hook: code=%d stderr=%q", code, stderr.String())
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("hook blocked for %v", elapsed)
	}
}
