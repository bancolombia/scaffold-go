package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateDaDatabaseCommand_Structure(t *testing.T) {
	cmd := NewGenerateDaDatabaseCommand()

	if cmd.Use != "database" {
		t.Errorf("expected Use to be 'database', got '%s'", cmd.Use)
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

func TestGenerateDaDatabaseCommand_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{
			name:     "entity flag is registered",
			flagName: "entity",
			wantType: "string",
		},
		{
			name:     "type flag is registered",
			flagName: "type",
			wantType: "string",
		},
		{
			name:     "fields flag is registered",
			flagName: "fields",
			wantType: "stringSlice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateDaDatabaseCommand()
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

func TestGenerateDaDatabaseCommand_Args(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "no args is valid",
			args:      []string{},
			wantError: false,
		},
		{
			name:      "one arg returns error",
			args:      []string{"UnexpectedPositional"},
			wantError: true,
		},
		{
			name:      "two args returns error",
			args:      []string{"UserRepository", "Extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateDaDatabaseCommand()

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

func TestGenerateDaDatabaseCommand_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateDaDatabaseCommand()
	cmd.SetArgs([]string{"--entity", "User", "--type", "postgres", "--fields", "Name:string"})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for uninitialized project, got none")
	}
}

func TestGenerateDaDatabaseCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()

	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := []byte("module test/project\n\ngo 1.21\n")
	if err := os.WriteFile(goModPath, goModContent, 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	dbDir := filepath.Join(tmpDir, "internal", "drivenadapters", "database")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("failed to create database directory: %v", err)
	}

	domainDir := filepath.Join(tmpDir, "internal", "domain")
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatalf("failed to create domain directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(domainDir, "user.go"), []byte("package domain\n"), 0644); err != nil {
		t.Fatalf("failed to seed entity file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(domainDir, "user_repository.go"), []byte("package domain\n"), 0644); err != nil {
		t.Fatalf("failed to seed repository file: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateDaDatabaseCommand()
	cmd.SetArgs([]string{"--entity", "User", "--type", "postgres", "--fields", "Name:string,Email:string"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	adapterPath := filepath.Join(dbDir, "postgres_adapter.go")
	if _, err := os.Stat(adapterPath); os.IsNotExist(err) {
		t.Error("postgres_adapter.go was not created")
	}

	repoPath := filepath.Join(dbDir, "user_repository.go")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Error("user_repository.go was not created")
	}

	testPath := filepath.Join(dbDir, "user_repository_test.go")
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Error("user_repository_test.go was not created")
	}

	modulePath := filepath.Join(dbDir, "module.go")
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		t.Error("module.go file was not created")
	}
}

func TestGenerateDaDatabaseCommand_Example(t *testing.T) {
	cmd := NewGenerateDaDatabaseCommand()

	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}

	if !contains(cmd.Example, "scaffold-go generate da database") {
		t.Error("example should contain 'scaffold-go generate da database'")
	}

	if !contains(cmd.Example, "--entity") {
		t.Error("example should contain '--entity' flag")
	}

	if !contains(cmd.Example, "--type") {
		t.Error("example should contain '--type' flag")
	}
}
