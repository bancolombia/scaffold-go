package generator

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleMainGo = `package main

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

func writeMain(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "cmd"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmd", "main.go"), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return dir
}

func readMain(t *testing.T, projectDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectDir, "cmd", "main.go"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(data)
}

func TestEnsureMainImport_AddsImportBetweenMarkers(t *testing.T) {
	dir := writeMain(t, sampleMainGo)

	if err := EnsureMainImport(dir, "github.com/test/project/internal/entrypoints/messaging", "consumers"); err != nil {
		t.Fatalf("EnsureMainImport returned error: %v", err)
	}

	body := readMain(t, dir)
	expected := "\tconsumers \"github.com/test/project/internal/entrypoints/messaging\""
	if !strings.Contains(body, expected) {
		t.Fatalf("import line not inserted; got body:\n%s", body)
	}
}

func TestEnsureMainImport_Idempotent(t *testing.T) {
	dir := writeMain(t, sampleMainGo)

	const importPath = "github.com/test/project/internal/entrypoints/messaging"
	if err := EnsureMainImport(dir, importPath, "consumers"); err != nil {
		t.Fatalf("first call returned error: %v", err)
	}
	if err := EnsureMainImport(dir, importPath, "consumers"); err != nil {
		t.Fatalf("second call returned error: %v", err)
	}

	body := readMain(t, dir)
	count := strings.Count(body, importPath)
	if count != 1 {
		t.Fatalf("expected import path to appear exactly once, got %d", count)
	}
}

func TestEnsureMainImport_MissingMarkersWarning(t *testing.T) {
	dir := writeMain(t, "package main\n\nimport \"go.uber.org/zap\"\n")

	err := EnsureMainImport(dir, "example/foo", "")
	if err == nil {
		t.Fatal("expected warning when markers are missing")
	}
	var w *MainWiringWarning
	if !errors.As(err, &w) {
		t.Fatalf("expected MainWiringWarning, got %T", err)
	}
}

func TestEnsureMainModule_AddsModuleBetweenMarkers(t *testing.T) {
	dir := writeMain(t, sampleMainGo)

	if err := EnsureMainModule(dir, "consumers.OrderConsumerModule"); err != nil {
		t.Fatalf("EnsureMainModule returned error: %v", err)
	}

	body := readMain(t, dir)
	expected := "\t\tconsumers.OrderConsumerModule,"
	if !strings.Contains(body, expected) {
		t.Fatalf("module line not inserted; got body:\n%s", body)
	}
}

func TestEnsureMainModule_Idempotent(t *testing.T) {
	dir := writeMain(t, sampleMainGo)

	const expr = "consumers.OrderConsumerModule"
	if err := EnsureMainModule(dir, expr); err != nil {
		t.Fatalf("first call returned error: %v", err)
	}
	if err := EnsureMainModule(dir, expr); err != nil {
		t.Fatalf("second call returned error: %v", err)
	}

	body := readMain(t, dir)
	count := strings.Count(body, expr)
	if count != 1 {
		t.Fatalf("expected module expression to appear exactly once, got %d", count)
	}
}
