package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateDaCacheCommand_Structure(t *testing.T) {
	cmd := NewGenerateDaCacheCommand()

	if cmd.Use != "cache [name]" {
		t.Errorf("expected Use to be 'cache [name]', got '%s'", cmd.Use)
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

func TestGenerateDaCacheCommand_Args(t *testing.T) {
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
			args:      []string{"UserCache"},
			wantError: false,
		},
		{
			name:      "two args returns error",
			args:      []string{"UserCache", "Extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateDaCacheCommand()
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

func TestGenerateDaCacheCommand_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateDaCacheCommand()
	cmd.SetArgs([]string{"UserCache"})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for uninitialized project, got none")
	}
}

func TestGenerateDaCacheCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()

	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := []byte("module test/project\n\ngo 1.21\n")
	if err := os.WriteFile(goModPath, goModContent, 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	cacheDir := filepath.Join(tmpDir, "internal", "drivenadapters", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache directory: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateDaCacheCommand()
	cmd.SetArgs([]string{"UserCache"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	cachePath := filepath.Join(cacheDir, "usercache_cache.go")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("cache file was not created")
	}
}

func TestGenerateDaCacheCommand_Example(t *testing.T) {
	cmd := NewGenerateDaCacheCommand()

	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}

	if !contains(cmd.Example, "scaffold-go generate da cache") {
		t.Error("example should contain 'scaffold-go generate da cache'")
	}
}
