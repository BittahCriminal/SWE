package swe

import "testing"

func TestParseModes(t *testing.T) {
	tests := []struct {
		input        string
		mode         Mode
		valid        bool
		validDefault bool
	}{
		{" LITE ", ModeLite, true, true},
		{"full", ModeFull, true, true},
		{"ULTRA", ModeUltra, true, true},
		{"off", ModeOff, true, true},
		{"review", ModeReview, true, false},
		{"fast", "", false, false},
	}
	for _, test := range tests {
		mode, valid := ParseMode(test.input)
		if mode != test.mode || valid != test.valid {
			t.Errorf("ParseMode(%q) = %q, %v; want %q, %v", test.input, mode, valid, test.mode, test.valid)
		}
		_, validDefault := ParseDefaultMode(test.input)
		if validDefault != test.validDefault {
			t.Errorf("ParseDefaultMode(%q) valid = %v; want %v", test.input, validDefault, test.validDefault)
		}
	}
	if _, ok := ParseRuntimeMode("review"); ok {
		t.Fatal("review must not be a persistent runtime mode")
	}
}

func TestDeactivationMustBeStandalone(t *testing.T) {
	for _, input := range []string{"stop swe", "STOP SWE!", " normal mode... "} {
		if !IsDeactivationCommand(input) {
			t.Errorf("expected %q to deactivate", input)
		}
	}
	for _, input := range []string{"add a normal mode toggle", "do not stop swe yet", "normal modes"} {
		if IsDeactivationCommand(input) {
			t.Errorf("did not expect %q to deactivate", input)
		}
	}
}
