package swe

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultCodebaseMapBudget = 2000
	maxCodebaseMapFiles      = 4000
	maxCodebaseMapRead       = 64 << 10
	maxGitFileList           = 64 << 20
	maxMapSymbolLength       = 128
	maxMapSymbolsPerFile     = 3
)

var (
	mapNamePattern = regexp.MustCompile(`^(?:type\s+)?(?:[\w$]+\s+as\s+)?([A-Za-z_$][\w$]*(?:\.[A-Za-z_$][\w$]*)?)$`)
	mapPatterns    = map[string][]*regexp.Regexp{
		".py": {
			regexp.MustCompile(`(?m)^(?:async\s+)?def\s+([A-Za-z]\w*)`),
			regexp.MustCompile(`(?m)^class\s+([A-Za-z]\w*)`),
		},
		".js": {
			regexp.MustCompile(`(?m)^export\s+(?:default\s+)?(?:async\s+)?(?:function\*?|class|const|let|var|interface|type|enum)\s+([A-Za-z_$][\w$]*)`),
			regexp.MustCompile(`(?m)^export\s*\{([^}]*)\}`),
		},
		".go": {
			regexp.MustCompile(`(?m)^func\s+(?:\([^)]*\)\s*)?([A-Z]\w*)`),
			regexp.MustCompile(`(?m)^type\s+([A-Z]\w*)`),
		},
		".rs": {regexp.MustCompile(`(?m)^pub\s+(?:async\s+)?(?:fn|struct|enum|trait|type)\s+(\w+)`)},
		".rb": {regexp.MustCompile(`(?m)^\s*(?:def|class|module)\s+([A-Za-z][\w.]*)`)},
		"jvm": {regexp.MustCompile(`(?m)^\s*(?:public\s+)?(?:static\s+)?(?:final\s+)?(?:class|interface|enum|record)\s+(\w+)`)},
	}
)

type cappedBuffer struct {
	bytes.Buffer
	max int
}

func (buffer *cappedBuffer) Write(data []byte) (int, error) {
	remaining := buffer.max - buffer.Len()
	if remaining > 0 {
		if remaining > len(data) {
			remaining = len(data)
		}
		_, _ = buffer.Buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func codebaseMapBudget() int {
	if value := strings.TrimSpace(os.Getenv("SWE_MAP_CHARS")); value != "" {
		if budget, err := strconv.Atoi(value); err == nil && budget >= 0 {
			return budget
		}
	}
	return defaultCodebaseMapBudget
}

// BuildCodebaseMap returns a bounded list of reusable top-level symbols. Only
// identifiers and allowlisted paths reach the prompt; symlinks, tests, build
// output, generated code, and unsafe filenames are ignored.
func BuildCodebaseMap(root string, budget int) string {
	if strings.TrimSpace(root) == "" || budget <= 0 {
		return ""
	}
	files := listCodebaseMapFiles(root)
	sort.Slice(files, func(i, j int) bool {
		iShared, jShared := isSharedMapPath(files[i]), isSharedMapPath(files[j])
		if iShared != jShared {
			return iShared
		}
		return files[i] < files[j]
	})
	if len(files) > maxCodebaseMapFiles {
		files = files[:maxCodebaseMapFiles]
	}

	type directorySymbols struct {
		dir     string
		perFile [][]string
	}
	directories := make([]directorySymbols, 0)
	indexes := map[string]int{}
	for _, name := range files {
		if !safeMapPath(name) || skipMapPath(name) || !mapSourceFile(name) {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(name))
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		opened, statErr := file.Stat()
		if statErr != nil || !os.SameFile(info, opened) {
			_ = file.Close()
			continue
		}
		data, readErr := io.ReadAll(io.LimitReader(file, maxCodebaseMapRead))
		_ = file.Close()
		if readErr != nil {
			continue
		}
		names := codebaseMapSymbols(name, string(data))
		if len(names) == 0 {
			continue
		}
		dir := filepath.ToSlash(filepath.Dir(name))
		index, ok := indexes[dir]
		if !ok {
			index = len(directories)
			indexes[dir] = index
			directories = append(directories, directorySymbols{dir: dir})
		}
		directories[index].perFile = append(directories[index].perFile, names)
	}

	lines := make([]string, 0, len(directories))
	used, skipped := 0, 0
	for _, directory := range directories {
		limit := 3
		if len(directory.perFile) > 8 {
			limit = 1
		}
		names := make([]string, 0)
		seen := map[string]bool{}
		for _, perFile := range directory.perFile {
			for _, name := range perFile[:min(limit, len(perFile))] {
				if !seen[name] {
					seen[name] = true
					names = append(names, name)
				}
			}
		}
		suffix := ""
		if len(names) > 30 {
			names = names[:30]
			suffix = ", ..."
		}
		line := directory.dir + "/: " + strings.Join(names, ", ") + suffix
		if used+len(line) > budget {
			skipped++
			continue
		}
		lines = append(lines, line)
		used += len(line) + 1
	}
	if len(lines) == 0 {
		return ""
	}
	result := "Codebase map (what already exists; reuse it, read a file only when you need its details):\n" + strings.Join(lines, "\n")
	if skipped > 0 {
		result += "\n(" + strconv.Itoa(skipped) + " more folders not listed)"
	}
	return result
}

func listCodebaseMapFiles(root string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	buffer := &cappedBuffer{max: maxGitFileList}
	command := exec.CommandContext(ctx, "git", "-C", root, "ls-files", "-z")
	command.Stdout = buffer
	command.Stderr = nil
	_ = command.Run()
	if buffer.Len() > 0 {
		parts := bytes.Split(buffer.Bytes(), []byte{0})
		files := make([]string, 0, min(len(parts), maxCodebaseMapFiles*5))
		for _, part := range parts {
			if len(part) == 0 {
				continue
			}
			files = append(files, filepath.ToSlash(string(part)))
			if len(files) >= maxCodebaseMapFiles*5 {
				break
			}
		}
		if len(files) > 0 {
			return files
		}
	}

	files := make([]string, 0)
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == "." {
			return nil
		}
		name := filepath.ToSlash(relative)
		depth := strings.Count(name, "/")
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") || depth >= 6 || skipMapDirectory(name) {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, name)
		if len(files) >= maxCodebaseMapFiles*5 {
			return fs.SkipAll
		}
		return nil
	})
	return files
}

func codebaseMapSymbols(name, text string) []string {
	extension := strings.ToLower(filepath.Ext(name))
	key := extension
	switch extension {
	case ".jsx", ".mjs", ".cjs", ".ts", ".tsx":
		key = ".js"
	case ".java", ".kt", ".php", ".swift", ".cs":
		key = "jvm"
	}
	patterns := mapPatterns[key]
	names := make([]string, 0)
	seen := map[string]bool{}
	for _, pattern := range patterns {
		for _, match := range pattern.FindAllStringSubmatch(text, -1) {
			for _, candidate := range strings.Split(match[1], ",") {
				parsed := mapNamePattern.FindStringSubmatch(strings.TrimSpace(candidate))
				if parsed == nil || len(parsed[1]) > maxMapSymbolLength || strings.HasPrefix(parsed[1], "_") || seen[parsed[1]] {
					continue
				}
				seen[parsed[1]] = true
				names = append(names, strings.Clone(parsed[1]))
				if len(names) == maxMapSymbolsPerFile {
					return names
				}
			}
		}
	}
	return names
}

func safeMapPath(name string) bool {
	if name == "" {
		return false
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if strings.HasPrefix(name, "/") || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || filepath.ToSlash(clean) != name {
		return false
	}
	for _, character := range name {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || strings.ContainsRune("_.@+$()[]/-", character) {
			continue
		}
		return false
	}
	return true
}

func mapSourceFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".py", ".js", ".jsx", ".mjs", ".cjs", ".ts", ".tsx", ".go", ".rs", ".rb", ".java", ".kt", ".php", ".swift", ".cs":
		return true
	default:
		return false
	}
}

func skipMapDirectory(name string) bool {
	for _, part := range strings.Split(strings.ToLower(filepath.ToSlash(name)), "/") {
		switch part {
		case "node_modules", ".git", "dist", "build", "out", "vendor", "venv", ".venv", "__pycache__", "coverage", ".next", "target", "migrations", "versions", "fixtures", "test", "tests", "__tests__", "spec":
			return true
		}
	}
	return false
}

func skipMapPath(name string) bool {
	if skipMapDirectory(filepath.ToSlash(filepath.Dir(name))) {
		return true
	}
	base := strings.ToLower(filepath.Base(name))
	extension := filepath.Ext(base)
	stem := strings.TrimSuffix(base, extension)
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(stem, "_test") ||
		strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") ||
		strings.Contains(base, ".gen.") || strings.Contains(base, ".min.") || strings.Contains(base, ".d.")
}

func isSharedMapPath(name string) bool {
	lower := strings.ToLower(name)
	for _, part := range []string{"util", "helper", "lib", "common", "shared", "service", "component", "hook", "core", "model", "schema", "api"} {
		if strings.Contains(lower, part) {
			return true
		}
	}
	return false
}
