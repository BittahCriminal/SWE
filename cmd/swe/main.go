package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	swe "github.com/bittahcriminal/SWE"
)

var version = "dev"
var hookInputTimeout = time.Second

const usage = `SWE — smallest-working-engineering mode for AI coding agents

Usage:
  swe instructions [lite|full|ultra|review]
  swe mode [lite|full|ultra|review|off]
  swe default [lite|full|ultra|off]
  swe status [--badge]
  swe install opencode [--config-dir PATH]
  swe hook session-start|prompt-submit|subagent-start
  swe version
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stdout, usage)
		return 0
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, version)
		return 0
	case "instructions":
		mode := swe.ResolveDefaultMode()
		if len(args) == 2 {
			var ok bool
			mode, ok = swe.ParseMode(args[1])
			if !ok || mode == swe.ModeOff {
				return usageError(stderr, "instructions mode must be lite, full, ultra, or review")
			}
		} else if len(args) != 1 {
			return usageError(stderr, "usage: swe instructions [lite|full|ultra|review]")
		}
		fmt.Fprint(stdout, swe.BuildInstructions(mode))
		return 0
	case "mode":
		if len(args) == 1 {
			if mode, ok := swe.ReadMode(); ok {
				fmt.Fprintln(stdout, mode)
			} else {
				fmt.Fprintln(stdout, swe.ModeOff)
			}
			return 0
		}
		if len(args) != 2 {
			return usageError(stderr, "usage: swe mode [lite|full|ultra|review|off]")
		}
		mode, ok := swe.ParseMode(args[1])
		if !ok {
			return usageError(stderr, "invalid mode: "+args[1])
		}
		if err := swe.SetMode(mode); err != nil {
			return runtimeError(stderr, err)
		}
		fmt.Fprintln(stdout, mode)
		return 0
	case "default":
		if len(args) == 1 {
			fmt.Fprintln(stdout, swe.ResolveDefaultMode())
			return 0
		}
		if len(args) != 2 {
			return usageError(stderr, "usage: swe default [lite|full|ultra|off]")
		}
		mode, ok := swe.ParseDefaultMode(args[1])
		if !ok {
			return usageError(stderr, "invalid default mode: "+args[1])
		}
		if err := swe.SaveDefaultMode(mode); err != nil {
			return runtimeError(stderr, err)
		}
		fmt.Fprintln(stdout, mode)
		return 0
	case "status":
		if len(args) == 2 && args[1] == "--badge" {
			fmt.Fprint(stdout, swe.Badge())
			return 0
		}
		if len(args) != 1 {
			return usageError(stderr, "usage: swe status [--badge]")
		}
		active := string(swe.ModeOff)
		if mode, ok := swe.ReadMode(); ok {
			active = string(mode)
		}
		fmt.Fprintf(stdout, "mode=%s default=%s\n", active, swe.ResolveDefaultMode())
		return 0
	case "install":
		const installUsage = "usage: swe install opencode [--config-dir PATH]"
		if len(args) < 2 || args[1] != "opencode" {
			return usageError(stderr, installUsage)
		}
		var configDir string
		switch {
		case len(args) == 4 && args[2] == "--config-dir" && strings.TrimSpace(args[3]) != "":
			configDir = args[3]
		case len(args) == 2:
			var err error
			if configDir, err = swe.OpenCodeConfigDir(); err != nil {
				return runtimeError(stderr, err)
			}
		default:
			return usageError(stderr, installUsage)
		}
		if err := swe.InstallOpenCode(configDir); err != nil {
			return runtimeError(stderr, err)
		}
		fmt.Fprintf(stdout, "SWE installed for OpenCode at %s\n", configDir)
		return 0
	case "hook":
		if len(args) != 2 {
			return usageError(stderr, "usage: swe hook session-start|prompt-submit|subagent-start")
		}
		var output []byte
		switch strings.ToLower(args[1]) {
		case "session-start":
			output = swe.HandleSessionStart()
		case "prompt-submit":
			output = runInputHook(stdin, swe.HandlePromptSubmit)
		case "subagent-start":
			output = runInputHook(stdin, swe.HandleSubagentStart)
		default:
			return usageError(stderr, "unknown hook: "+args[1])
		}
		_, _ = stdout.Write(output)
		return 0
	default:
		return usageError(stderr, "unknown command: "+args[0])
	}
}

func runInputHook(stdin io.Reader, hook func(io.Reader) []byte) []byte {
	result := make(chan []byte, 1)
	go func() { result <- hook(stdin) }()
	timer := time.NewTimer(hookInputTimeout)
	defer timer.Stop()
	select {
	case output := <-result:
		return output
	case <-timer.C:
		return nil
	}
}

func usageError(stderr io.Writer, message string) int {
	fmt.Fprintln(stderr, message)
	return 2
}

func runtimeError(stderr io.Writer, err error) int {
	fmt.Fprintln(stderr, "swe:", err)
	return 1
}
