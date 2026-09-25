package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateEpHttpCommand_Structure(t *testing.T) {
	cmd := NewGenerateEpHttpCommand()

	if cmd.Use != "http [name]" {
		t.Errorf("expected Use to be 'http [name]', got '%s'", cmd.Use)
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

func TestGenerateEpHttpCommand_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{
			name:     "usecases flag is registered",
			flagName: "usecases",
			wantType: "stringSlice",
		},
		{
			name:     "routes flag is registered",
			flagName: "routes",
			wantType: "stringSlice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateEpHttpCommand()
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

func TestGenerateEpHttpCommand_Args(t *testing.T) {
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
			args:      []string{"UserHandler"},
			wantError: false,
		},
		{
			name:      "two args returns error",
			args:      []string{"UserHandler", "Extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateEpHttpCommand()
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

func TestGenerateEpHttpCommand_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateEpHttpCommand()
	cmd.SetArgs([]string{"UserHandler"})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for uninitialized project, got none")
	}
}

func TestGenerateEpHttpCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()

	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := []byte("module test/project\n\ngo 1.21\n")
	if err := os.WriteFile(goModPath, goModContent, 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	httpDir := filepath.Join(tmpDir, "internal", "entrypoints", "http")
	if err := os.MkdirAll(httpDir, 0755); err != nil {
		t.Fatalf("failed to create http directory: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateEpHttpCommand()
	cmd.SetArgs([]string{"UserHandler", "--usecases", "createUser:CreateUserUseCase", "--routes", "POST:/users:CreateUser"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	handlerPath := filepath.Join(httpDir, "handler", "user_handler.go")
	if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
		t.Error("handler file was not created")
	}
}

func TestGenerateEpHttpCommand_Example(t *testing.T) {
	cmd := NewGenerateEpHttpCommand()

	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}

	if !contains(cmd.Example, "scaffold-go generate ep http") {
		t.Error("example should contain 'scaffold-go generate ep http'")
	}

	if !contains(cmd.Example, "--usecases") {
		t.Error("example should contain '--usecases' flag")
	}

	if !contains(cmd.Example, "--routes") {
		t.Error("example should contain '--routes' flag")
	}
}
