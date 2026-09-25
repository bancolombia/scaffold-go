package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateUseCaseCommand_Structure(t *testing.T) {
	cmd := NewGenerateUseCaseCommand()

	if cmd.Use != "usecase [name]" {
		t.Errorf("expected Use to be 'usecase [name]', got '%s'", cmd.Use)
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

func TestGenerateUseCaseCommand_Flags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantType string
	}{
		{
			name:     "dependencies flag is registered",
			flagName: "dependencies",
			wantType: "stringSlice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateUseCaseCommand()
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

func TestGenerateUseCaseCommand_Args(t *testing.T) {
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
			args:      []string{"CreateUser"},
			wantError: false,
		},
		{
			name:      "two args returns error",
			args:      []string{"CreateUser", "Extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGenerateUseCaseCommand()
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

func TestGenerateUseCaseCommand_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateUseCaseCommand()
	cmd.SetArgs([]string{"CreateUser"})

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for uninitialized project, got none")
	}
}

func TestGenerateUseCaseCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()

	goModPath := filepath.Join(tmpDir, "go.mod")
	goModContent := []byte("module test/project\n\ngo 1.21\n")
	if err := os.WriteFile(goModPath, goModContent, 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	usecaseDir := filepath.Join(tmpDir, "internal", "usecase")
	if err := os.MkdirAll(usecaseDir, 0755); err != nil {
		t.Fatalf("failed to create usecase directory: %v", err)
	}

	if err := writeTestMainFile(tmpDir, "test/project"); err != nil {
		t.Fatalf("failed to create cmd/main.go: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewGenerateUseCaseCommand()
	cmd.SetArgs([]string{"CreateUser", "--dependencies", "userRepo:UserRepository"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	usecasePath := filepath.Join(usecaseDir, "createuser_usecase.go")
	if _, err := os.Stat(usecasePath); os.IsNotExist(err) {
		t.Error("usecase file was not created")
	}

	modulePath := filepath.Join(usecaseDir, "module.go")
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		t.Error("usecase module.go file was not created")
	}

	mainPath := filepath.Join(tmpDir, "cmd", "main.go")
	mainContent, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	if !strings.Contains(string(mainContent), "test/project/internal/usecase") {
		t.Error("main.go was not updated with usecase import")
	}
	if !strings.Contains(string(mainContent), "usecase.Module") {
		t.Error("main.go was not updated with usecase.Module")
	}
}

// writeTestMainFile creates a minimal cmd/main.go file that the use case
// generator can update via AST manipulation (it expects fx.New, fx.Invoke,
// and an import of "<modulePath>/config").
func writeTestMainFile(projectDir, modulePath string) error {
	cmdDir := filepath.Join(projectDir, "cmd")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}

	content := `package main

import (
	"context"

	"` + modulePath + `/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			zap.NewProduction,
		),
		fx.Invoke(run),
	)

	app.Run()
}

func run(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting application")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping application")
			return nil
		},
	})
}
`
	return os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(content), 0644)
}

func TestGenerateUseCaseCommand_Example(t *testing.T) {
	cmd := NewGenerateUseCaseCommand()

	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}

	if !contains(cmd.Example, "scaffold-go generate usecase") {
		t.Error("example should contain 'scaffold-go generate usecase'")
	}

	if !contains(cmd.Example, "--dependencies") {
		t.Error("example should contain '--dependencies' flag")
	}
}
