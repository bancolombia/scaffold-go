package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bancolombia/scaffold-go/internal/template"
)

// ConfigWiringWarning is returned by EnsureConfigBlock when
// config/config.yaml cannot be updated automatically (e.g. the file has been
// removed or renamed). Mirrors MainWiringWarning so callers can decide whether
// to print a friendly hint or surface a hard error.
type ConfigWiringWarning struct {
	Path    string
	Message string
}

func (w *ConfigWiringWarning) Error() string {
	return fmt.Sprintf("config/config.yaml: %s", w.Message)
}

// EnsureConfigBlock idempotently appends a YAML block to
// <projectDir>/config/config.yaml. It uses a line-oriented, column-0 scan for
// `<key>:` to decide whether the block is already there. When absent, the
// rendered snippet is appended at end-of-file with a single blank line of
// separation from the previous content.
//
// Extension recipe for future generators (broker / cache / api / …):
//  1. Add internal/template/templates/config/<key>.yaml.tmpl (top-level YAML
//     block whose first line is `<key>:`).
//  2. Call EnsureConfigBlock(projectDir, "<key>", rendered) after WriteFiles
//     in the generator's RunE.
//  3. Add a test asserting the block is present + a re-run idempotency test.
func EnsureConfigBlock(projectDir, key, snippet string) (*ConfigWiringWarning, error) {
	configPath := filepath.Join(projectDir, "config", "config.yaml")
	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigWiringWarning{
				Path:    configPath,
				Message: "file not found; skipping config injection",
			}, nil
		}
		return nil, fmt.Errorf("stat config.yaml: %w", err)
	}

	body, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config.yaml: %w", err)
	}

	if hasTopLevelKey(body, key) {
		return nil, nil
	}

	updated := appendBlock(body, snippet)
	if err := os.WriteFile(configPath, updated, info.Mode().Perm()); err != nil {
		return nil, fmt.Errorf("write config.yaml: %w", err)
	}
	return nil, nil
}

// EnsureConfigBlockForDatabase renders the config snippet template for the
// given database type and appends the resulting block to config/config.yaml.
// The template is looked up under `config/<dbType>.yaml.tmpl` in the template
// engine's registry. Returns (nil, nil) when the block is already present or
// when the database type has no config snippet template.
func EnsureConfigBlockForDatabase(projectDir, dbType, projectName string, engine *template.TemplateEngine) (*ConfigWiringWarning, error) {
	templateName := "config/" + dbType + ".yaml.tmpl"
	snippet, err := engine.Execute(templateName, map[string]any{"Name": projectName})
	if err != nil {
		// No snippet template registered for this dbType — this is a soft
		// no-op (e.g. postgres, which is covered by the base config).
		if strings.Contains(err.Error(), "template not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("render config snippet %q: %w", templateName, err)
	}
	return EnsureConfigBlock(projectDir, dbType, snippet)
}

// hasTopLevelKey reports whether `body` contains a line whose first
// non-newline characters are `<key>:`. This is a column-0 match, so keys
// nested under other maps or occurring inside string values do not count.
func hasTopLevelKey(body []byte, key string) bool {
	needle := []byte(key + ":")
	for _, raw := range bytes.Split(body, []byte("\n")) {
		if len(raw) < len(needle) {
			continue
		}
		if !bytes.HasPrefix(raw, needle) {
			continue
		}
		// Ensure the character right after `<key>:` is either end-of-line,
		// whitespace, or a comment marker — so `mongodbfoo:` doesn't count.
		after := raw[len(needle):]
		if len(after) == 0 {
			return true
		}
		switch after[0] {
		case ' ', '\t', '\r', '#':
			return true
		}
	}
	return false
}

// appendBlock returns `body` with `snippet` appended at end-of-file, with
// exactly one blank line separating the two and a trailing newline.
func appendBlock(body []byte, snippet string) []byte {
	trimmedBody := bytes.TrimRight(body, "\n")
	trimmedSnippet := strings.TrimRight(snippet, "\n")

	var buf bytes.Buffer
	if len(trimmedBody) > 0 {
		buf.Write(trimmedBody)
		buf.WriteString("\n\n")
	}
	buf.WriteString(trimmedSnippet)
	buf.WriteString("\n")
	return buf.Bytes()
}
