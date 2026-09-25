package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Validator struct{}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`).MatchString(name) {
		return fmt.Errorf("name must start with uppercase letter and contain only alphanumeric characters")
	}
	return nil
}

func (v *Validator) ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(name) {
		return fmt.Errorf("project name must start with lowercase letter and contain only lowercase letters, numbers, and hyphens")
	}
	return nil
}

func (v *Validator) ValidateFields(fields []string) error {
	for _, field := range fields {
		parts := strings.Split(field, ":")
		if len(parts) != 2 {
			return fmt.Errorf("field must be in format 'name:type', got: %s", field)
		}
		name, typ := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if name == "" || typ == "" {
			return fmt.Errorf("field name and type cannot be empty: %s", field)
		}
		if !regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`).MatchString(name) {
			return fmt.Errorf("field name must start with uppercase letter: %s", name)
		}
	}
	return nil
}

func (v *Validator) ValidateDirectoryNotExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("directory already exists: %s", path)
	}
	return nil
}

func (v *Validator) ValidateDirectoryExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("directory does not exist: %s", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	return nil
}

func (v *Validator) FileExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file does not exist: %s", path)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}
	return nil
}

func (v *Validator) ValidateProjectInitialized(dir string) error {
	// Walk up the directory tree to find go.mod
	currentDir := dir
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// go.mod found
			return nil
		}

		// Get parent directory
		parent := filepath.Join(currentDir, "..")
		absParent, err := filepath.Abs(parent)
		if err != nil {
			break
		}

		// Check if we've reached the root
		absCurrent, err := filepath.Abs(currentDir)
		if err != nil || absParent == absCurrent {
			break
		}

		currentDir = absParent
	}

	return fmt.Errorf("project not initialized: no go.mod file found in current or parent directories. Run 'scaffold-go init <project-name>' first")
}

func (v *Validator) GetModulePath(dir string) (string, error) {
	// Walk up the directory tree to find go.mod
	currentDir := dir
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// go.mod found, read it
			content, err := os.ReadFile(goModPath)
			if err != nil {
				return "", fmt.Errorf("failed to read go.mod: %w", err)
			}

			// Parse module path from first line
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					modulePath := strings.TrimSpace(strings.TrimPrefix(line, "module"))
					return modulePath, nil
				}
			}
			return "", fmt.Errorf("module declaration not found in go.mod")
		}

		// Get parent directory
		parent := filepath.Join(currentDir, "..")
		absParent, err := filepath.Abs(parent)
		if err != nil {
			break
		}

		// Check if we've reached the root
		absCurrent, err := filepath.Abs(currentDir)
		if err != nil || absParent == absCurrent {
			break
		}

		currentDir = absParent
	}

	return "", fmt.Errorf("go.mod file not found in current or parent directories")
}
