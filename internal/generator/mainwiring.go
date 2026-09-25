package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Markers in cmd/main.go that fence the regions we patch idempotently.
const (
	importsBeginMarker = "// scaffold-go:imports:begin"
	importsEndMarker   = "// scaffold-go:imports:end"
	modulesBeginMarker = "// scaffold-go:modules:begin"
	modulesEndMarker   = "// scaffold-go:modules:end"
)

// MainWiringWarning is returned by EnsureMainImport / EnsureMainModule when
// cmd/main.go cannot be updated automatically (e.g. the markers are missing
// because the user pre-existed the project before this refactor). It lets the
// caller decide whether to print a friendly hint or surface a hard error.
type MainWiringWarning struct {
	Path    string
	Message string
}

func (w *MainWiringWarning) Error() string {
	return fmt.Sprintf("cmd/main.go: %s", w.Message)
}

// EnsureMainImport idempotently adds an import line inside the
// `// scaffold-go:imports:begin/end` block of `<projectDir>/cmd/main.go`.
//
// `alias` is optional; pass "" to import without an alias. If the import is
// already present (with or without the alias) this is a no-op.
//
// Returns a *MainWiringWarning when cmd/main.go cannot be patched (markers
// missing or file absent), so callers can decide between warning vs failing.
func EnsureMainImport(projectDir, importPath, alias string) error {
	mainPath := filepath.Join(projectDir, "cmd", "main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return &MainWiringWarning{Path: mainPath, Message: fmt.Sprintf("cannot read file: %v", err)}
	}

	var importLine string
	if alias != "" {
		importLine = fmt.Sprintf("\t%s %q", alias, importPath)
	} else {
		importLine = fmt.Sprintf("\t%q", importPath)
	}

	body := string(content)

	// Already wired?
	if strings.Contains(body, importLine) {
		return nil
	}
	// Already imported with/without an alias on a single line?
	if importedAny(body, importPath) {
		return nil
	}

	updated, ok := insertBetweenMarkers(body, importsBeginMarker, importsEndMarker, importLine)
	if !ok {
		return &MainWiringWarning{
			Path:    mainPath,
			Message: fmt.Sprintf("import markers not found; please add %s to the imports manually", importLine),
		}
	}

	if err := os.WriteFile(mainPath, []byte(updated), 0o644); err != nil {
		return &MainWiringWarning{Path: mainPath, Message: fmt.Sprintf("cannot write file: %v", err)}
	}
	return nil
}

// EnsureMainModule idempotently adds a module reference (e.g. `consumers.UserModule,`)
// inside the `// scaffold-go:modules:begin/end` block of `<projectDir>/cmd/main.go`.
//
// `expr` should be the module expression without a trailing comma; this helper
// appends one. If the same expression is already present this is a no-op.
func EnsureMainModule(projectDir, expr string) error {
	mainPath := filepath.Join(projectDir, "cmd", "main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return &MainWiringWarning{Path: mainPath, Message: fmt.Sprintf("cannot read file: %v", err)}
	}

	body := string(content)
	exprLine := fmt.Sprintf("\t\t%s,", expr)

	// Already wired?
	if strings.Contains(body, exprLine) {
		return nil
	}

	updated, ok := insertBetweenMarkers(body, modulesBeginMarker, modulesEndMarker, exprLine)
	if !ok {
		return &MainWiringWarning{
			Path:    mainPath,
			Message: fmt.Sprintf("module markers not found; please add %s to fx.New manually", exprLine),
		}
	}

	if err := os.WriteFile(mainPath, []byte(updated), 0o644); err != nil {
		return &MainWiringWarning{Path: mainPath, Message: fmt.Sprintf("cannot write file: %v", err)}
	}
	return nil
}

// insertBetweenMarkers inserts `line` immediately before the end marker of the
// fenced block. Returns the rewritten body and true on success, or the
// original body and false if the markers cannot be found.
func insertBetweenMarkers(body, beginMarker, endMarker, line string) (string, bool) {
	beginIdx := strings.Index(body, beginMarker)
	if beginIdx < 0 {
		return body, false
	}
	endIdx := strings.Index(body[beginIdx:], endMarker)
	if endIdx < 0 {
		return body, false
	}
	endIdx += beginIdx

	// Walk back to the start of the line containing the end marker so we can
	// insert the new entry on the line immediately above it, preserving the
	// indentation of the marker itself.
	lineStart := strings.LastIndex(body[:endIdx], "\n")
	if lineStart < 0 {
		lineStart = 0
	} else {
		lineStart++ // move past the '\n' so we keep the leading indentation of the marker line
	}

	return body[:lineStart] + line + "\n" + body[lineStart:], true
}

// importedAny returns true if `body` already contains an import line targeting
// `importPath`, regardless of whether it has an alias.
func importedAny(body, importPath string) bool {
	quoted := fmt.Sprintf("%q", importPath)
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if !strings.HasSuffix(line, quoted) {
			continue
		}
		// Either the whole line is the quoted path (no alias) or the prefix is
		// `<alias> ` then the quoted path.
		if line == quoted {
			return true
		}
		if strings.HasSuffix(line, " "+quoted) {
			return true
		}
	}
	return false
}
