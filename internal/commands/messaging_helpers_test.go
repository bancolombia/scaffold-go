package commands

import (
	"os"
	"path/filepath"
	"testing"
)

// newInitializedTempProject creates a temp directory that looks like a
// scaffold-go-initialized project (contains go.mod and a cmd/main.go with the
// scaffold-go markers wired up so EnsureMainImport / EnsureMainModule can patch it).
//
// Returns the absolute path to the temp directory.
func newInitializedTempProject(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test/project\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	dirs := []string{
		filepath.Join(tmpDir, "cmd"),
		filepath.Join(tmpDir, "internal", "entrypoints", "messaging"),
		filepath.Join(tmpDir, "internal", "drivenadapters", "messaging"),
		filepath.Join(tmpDir, "internal", "infrastructure", "messaging"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	mainGo := `package main

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	// scaffold-go:imports:begin
	// scaffold-go:imports:end
)

func main() {
	app := fx.New(
		fx.Provide(zap.NewProduction),
		// scaffold-go:modules:begin
		// scaffold-go:modules:end
		fx.Invoke(run),
	)
	app.Run()
}

func run(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return nil },
		OnStop:  func(ctx context.Context) error { return nil },
	})
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "cmd", "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	return tmpDir
}
