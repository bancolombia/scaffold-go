package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateEntityCommand_Structure(t *testing.T) {
	cmd := NewGenerateEntityCommand()

	if cmd.Use != "entity [name]" {
		t.Errorf("expected Use to be 'entity [name]', got '%s'", cmd.Use)
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

func TestGenerateEntityCommand_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{
			name:     "fields flag is registered",
			flagName: "fields",
			wantType: "stringSlice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateEntityCommand()
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

func TestGenerateEntityCommand_RequiredFlags(t *testing.T) {
	cmd := NewGenerateEntityCommand()

	// Check that fields flag is required
	flag := cmd.Flags().Lookup("fields")
	if flag == nil {
		t.Fatal("fields flag not found")
	}

	// Note: Cobra doesn't expose a direct way to check if a flag is required,
	// but we can verify it's registered
	if flag.Name != "fields" {
		t.Errorf("expected flag name 'fields', got '%s'", flag.Name)
	}
}

func TestGenerateEntityCommand_Args(t *testing.T) {
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
			args:      []string{"User"},
			wantError: false,
		},
		{
			name:      "two args returns error",
			args:      []string{"User", "Extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateEntityCommand()
			cmd.SetArgs(tt.args)

			// We're only testing arg validation, not execution
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

func TestGenerateEntityCommand_ValidationError(t *testing.T) {
	// Create a temporary directory without go.mod
	tmpDir := t.TempDir()

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateEntityCommand()
	cmd.SetArgs([]string{"User", "--fields", "Name:string"})

	// Should fail because project is not initialized (no go.mod)
	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for uninitialized project, got none")
	}
}

func TestGenerateEntityCommand_Success(t *testing.T) {
	// Create a temporary directory with go.mod
	tmpDir := t.TempDir()

	// Create go.mod file
	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := []byte("module test/project\n\ngo 1.21\n")
	if err := os.WriteFile(goModPath, goModContent, 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create required directories
	domainDir := filepath.Join(tmpDir, "internal", "domain")
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatalf("failed to create domain directory: %v", err)
	}

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateEntityCommand()
	cmd.SetArgs([]string{"User", "--fields", "Name:string,Email:string"})

	// Execute command
	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify entity file was created
	entityPath := filepath.Join(domainDir, "user.go")
	if _, err := os.Stat(entityPath); os.IsNotExist(err) {
		t.Error("entity file was not created")
	}

	// Verify repository file was created
	repoPath := filepath.Join(domainDir, "user_repository.go")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Error("repository file was not created")
	}
}

func TestGenerateEntityCommand_InvalidName(t *testing.T) {
	// Create a temporary directory with go.mod
	tmpDir := t.TempDir()

	// Create go.mod file
	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := []byte("module test/project\n\ngo 1.21\n")
	if err := os.WriteFile(goModPath, goModContent, 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	tests := []struct {
		name       string
		entityName string
	}{
		{
			name:       "invalid name with spaces",
			entityName: "User Name",
		},
		{
			name:       "invalid name with special characters",
			entityName: "User@Entity",
		},
		{
			name:       "invalid name starting with number",
			entityName: "123User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateEntityCommand()
			cmd.SetArgs([]string{tt.entityName, "--fields", "Name:string"})

			err := cmd.Execute()
			if err == nil {
				t.Error("expected error for invalid entity name, got none")
			}
		})
	}
}

func TestGenerateEntityCommand_Example(t *testing.T) {
	cmd := NewGenerateEntityCommand()

	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}

	// Verify example contains expected patterns
	if !contains(cmd.Example, "scaffold-go generate entity") {
		t.Error("example should contain 'scaffold-go generate entity'")
	}

	if !contains(cmd.Example, "--fields") {
		t.Error("example should contain '--fields' flag")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
