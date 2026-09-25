package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateDaApiCommand_Structure(t *testing.T) {
	cmd := NewGenerateDaApiCommand()

	if cmd.Use != "api [name]" {
		t.Errorf("expected Use to be 'api [name]', got '%s'", cmd.Use)
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

func TestGenerateDaApiCommand_Args(t *testing.T) {
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
			args:      []string{"PaymentService"},
			wantError: false,
		},
		{
			name:      "two args returns error",
			args:      []string{"PaymentService", "Extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateDaApiCommand()
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

func TestGenerateDaApiCommand_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateDaApiCommand()
	cmd.SetArgs([]string{"PaymentService"})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for uninitialized project, got none")
	}
}

func TestGenerateDaApiCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()

	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := []byte("module test/project\n\ngo 1.21\n")
	if err := os.WriteFile(goModPath, goModContent, 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	apiDir := filepath.Join(tmpDir, "internal", "drivenadapters", "api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		t.Fatalf("failed to create api directory: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateDaApiCommand()
	cmd.SetArgs([]string{"PaymentService"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	clientPath := filepath.Join(apiDir, "paymentservice_client.go")
	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		t.Error("api client file was not created")
	}
}

func TestGenerateDaApiCommand_Example(t *testing.T) {
	cmd := NewGenerateDaApiCommand()

	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}

	if !contains(cmd.Example, "scaffold-go generate da api") {
		t.Error("example should contain 'scaffold-go generate da api'")
	}
}
