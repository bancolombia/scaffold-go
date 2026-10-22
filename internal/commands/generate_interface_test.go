package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateInterfaceCommand_Structure(t *testing.T) {
	cmd := NewGenerateInterfaceCommand()

	if cmd.Use != "interface [name]" {
		t.Errorf("expected Use to be 'interface [name]', got '%s'", cmd.Use)
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
	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}
}

func TestGenerateInterfaceCommand_Flags(t *testing.T) {
	cmd := NewGenerateInterfaceCommand()

	entity := cmd.Flags().Lookup("entity")
	if entity == nil {
		t.Fatal("entity flag not found")
	}
	if entity.Shorthand != "e" {
		t.Errorf("entity shorthand = %q, want e", entity.Shorthand)
	}
	if entity.Value.Type() != "string" {
		t.Errorf("entity type = %q, want string", entity.Value.Type())
	}

	operation := cmd.Flags().Lookup("operation")
	if operation == nil {
		t.Fatal("operation flag not found")
	}
	if operation.Value.Type() != "string" {
		t.Errorf("operation type = %q, want string", operation.Value.Type())
	}
}

func TestGenerateInterfaceCommand_Args(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{name: "no args returns error", args: []string{}, wantError: true},
		{name: "one arg is valid", args: []string{"ProductReceiver"}, wantError: false},
		{name: "two args returns error", args: []string{"ProductReceiver", "Extra"}, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateInterfaceCommand()
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

func TestGenerateInterfaceCommand_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateInterfaceCommand()
	cmd.SetArgs([]string{"ProductReceiver", "--entity", "Product", "--operation", "ProcessProduct"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for uninitialized project, got none")
	}
}

func TestGenerateInterfaceCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test/project\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateInterfaceCommand()
	cmd.SetArgs([]string{"ProductReceiver", "--entity", "Product", "--operation", "ProcessProduct"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(tmpDir, "internal", "domain", "productreceiver.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("interface file was not created: %v", err)
	}
	text := string(content)
	for _, check := range []string{
		"type IProductReceiver interface",
		"ProcessProduct(ctx context.Context, product *Product) error",
	} {
		if !strings.Contains(text, check) {
			t.Errorf("file missing %q:\n%s", check, text)
		}
	}
}

func TestGenerateInterfaceCommand_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test/project\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateInterfaceCommand()
	cmd.SetArgs([]string{"IProductReceiver", "--entity", "Product", "--operation", "ProcessProduct"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for I-prefixed name, got none")
	}
}
