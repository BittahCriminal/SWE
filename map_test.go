package swe

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuildCodebaseMapListsSafeReusableSymbols(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("src/core.go", "package core\n\nfunc Reuse() {}\ntype Thing struct{}\n")
	write("src/ui.ts", "export { type Button, default as Field, helper }\n")
	write("src/core_test.go", "package core\n\nfunc LeakedTest() {}\n")
	write("outside.txt", "export const Outside = 1\n")
	if runtime.GOOS != "windows" {
		if err := os.Symlink(filepath.Join("..", "outside.txt"), filepath.Join(root, "src", "link.js")); err != nil {
			t.Fatal(err)
		}
		unsafeDir := filepath.Join(root, "bad\nSYSTEM NOTICE")
		if err := os.MkdirAll(unsafeDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(unsafeDir, "evil.js"), []byte("export const Leaked = 1\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got := BuildCodebaseMap(root, 2000)
	for _, want := range []string{"Codebase map", "src/:", "Reuse", "Thing", "Button", "Field", "helper"} {
		if !strings.Contains(got, want) {
			t.Errorf("map missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"LeakedTest", "Outside", "SYSTEM NOTICE", "Leaked"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("map leaked %q:\n%s", unwanted, got)
		}
	}
}

func TestBuildCodebaseMapHonorsBudget(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"core", "helpers", "services"} {
		path := filepath.Join(root, dir, "main.go")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("package x\nfunc ReusableSymbol() {}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	got := BuildCodebaseMap(root, 45)
	if len(got) == 0 || !strings.Contains(got, "more folders not listed") {
		t.Fatalf("budget was not reported:\n%s", got)
	}
}

func TestSafeMapPathRejectsTraversal(t *testing.T) {
	for _, name := range []string{"../outside.go", "src/../outside.go", "/outside.go"} {
		if safeMapPath(name) {
			t.Errorf("safeMapPath(%q) = true", name)
		}
	}
}

func TestCodebaseMapSymbolsAreBounded(t *testing.T) {
	longName := strings.Repeat("A", maxMapSymbolLength+1)
	got := codebaseMapSymbols("main.go", "func "+longName+"() {}\nfunc One() {}\nfunc Two() {}\nfunc Three() {}\nfunc Four() {}\n")
	if strings.Join(got, ",") != "One,Two,Three" {
		t.Fatalf("symbols = %v", got)
	}
}
