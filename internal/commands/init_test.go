package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand_Structure(t *testing.T) {
	cmd := NewInitCommand()

	if cmd.Use != "init [project-name]" {
		t.Errorf("expected Use to be 'init [project-name]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE function to be set")
	}
}

func TestInitCommand_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{
			name:     "module flag is registered",
			flagName: "module",
			wantType: "string",
		},
		{
			name:     "output flag is registered",
			flagName: "output",
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewInitCommand()
			flag := cmd.Flags().Lookup(tt.flagName)

			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			if flag.Value.Type() != tt.wantType {
				t.Errorf("flag %q type = %q, want %q", tt.flagName, flag.Value.Type(), tt.wantType)
			}
		})
	}
}

func TestInitCommand_Args(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "no args returns error",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "one arg is valid",
			args:      []string{"my-project"},
			wantError: false,
		},
		{
			name:      "two args returns error",
			args:      []string{"my-project", "extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewInitCommand()
			cmd.SetArgs(tt.args)

			err := cmd.Args(cmd, tt.args)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestInitCommand_InvalidProjectName(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		projectName string
	}{
		{
			name:        "project name with spaces",
			projectName: "my project",
		},
		{
			name:        "project name with special characters",
			projectName: "my@project",
		},
		{
			name:        "project name starting with number",
			projectName: "123project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewInitCommand()
			cmd.SetArgs([]string{tt.projectName, "--output", tmpDir})

			err := cmd.Execute()
			if err == nil {
				t.Error("expected error for invalid project name, got none")
			}
		})
	}
}

func TestInitCommand_DirectoryAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with the project name
	projectDir := filepath.Join(tmpDir, "existing-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"existing-project", "--output", tmpDir})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for existing directory, got none")
	}
}

func TestInitCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"test-project", "--module", "github.com/test/project", "--output", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	projectDir := filepath.Join(tmpDir, "test-project")

	// Verify go.mod was created
	goModPath := filepath.Join(projectDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("go.mod file was not created")
	}

	// Verify main.go was created
	mainGoPath := filepath.Join(projectDir, "cmd", "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Error("main.go file was not created")
	}

	// Verify config directory was created
	configDir := filepath.Join(projectDir, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("config directory was not created")
	}

	// Verify README.md was created
	readmePath := filepath.Join(projectDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md file was not created")
	}

	// Verify Makefile was created
	makefilePath := filepath.Join(projectDir, "Makefile")
	if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
		t.Error("Makefile was not created")
	}

	// Verify domain directory was created
	domainDir := filepath.Join(projectDir, "internal", "domain")
	if _, err := os.Stat(domainDir); os.IsNotExist(err) {
		t.Error("domain directory was not created")
	}

	// Verify entrypoints directories were created
	httpDir := filepath.Join(projectDir, "internal", "entrypoints", "http")
	if _, err := os.Stat(httpDir); os.IsNotExist(err) {
		t.Error("http entrypoints directory was not created")
	}

	// Verify drivenadapters directories were created
	dbDir := filepath.Join(projectDir, "internal", "drivenadapters", "database")
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		t.Error("database drivenadapters directory was not created")
	}
}

func TestInitCommand_DefaultModulePath(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"test-project", "--output", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	projectDir := filepath.Join(tmpDir, "test-project")

	// Read go.mod and verify default module path
	goModPath := filepath.Join(projectDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}

	goModContent := string(content)
	if !contains(goModContent, "module test-project") {
		t.Error("go.mod does not contain default module path")
	}
}

func TestInitCommand_Example(t *testing.T) {
	cmd := NewInitCommand()

	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}

	if !contains(cmd.Example, "scaffold-go init") {
		t.Error("example should contain 'scaffold-go init'")
	}

	if !contains(cmd.Example, "--module") {
		t.Error("example should contain '--module' flag")
	}
}

func TestInitCommand_HelpText(t *testing.T) {
	cmd := NewInitCommand()

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Verify help text mentions key concepts
	if !contains(cmd.Long, "clean architecture") && !contains(cmd.Short, "clean architecture") {
		t.Error("help text should mention clean architecture")
	}
}
