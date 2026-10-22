package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModulePathParser handles parsing of go.mod files
type ModulePathParser struct {
	projectDir string
}

// NewModulePathParser creates a new module path parser
func NewModulePathParser(projectDir string) *ModulePathParser {
	return &ModulePathParser{
		projectDir: projectDir,
	}
}

// Parse reads the go.mod file and extracts the module path
// Returns error if go.mod doesn't exist or module path cannot be parsed
func (p *ModulePathParser) Parse() (string, error) {
	goModPath := filepath.Join(p.projectDir, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			if modulePath == "" {
				return "", fmt.Errorf("empty module path in go.mod")
			}
			return modulePath, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading go.mod: %w", err)
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}
