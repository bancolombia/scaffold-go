package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
)

func TestInterfaceGenerator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		ifaceName string
		entity    string
		operation string
		wantError bool
	}{
		{
			name:      "valid names",
			ifaceName: "ProductReceiver",
			entity:    "Product",
			operation: "ProcessProduct",
			wantError: false,
		},
		{
			name:      "empty interface name",
			ifaceName: "",
			entity:    "Product",
			operation: "ProcessProduct",
			wantError: true,
		},
		{
			name:      "lowercase interface name",
			ifaceName: "productReceiver",
			entity:    "Product",
			operation: "ProcessProduct",
			wantError: true,
		},
		{
			name:      "interface name already has I prefix",
			ifaceName: "IProductReceiver",
			entity:    "Product",
			operation: "ProcessProduct",
			wantError: true,
		},
		{
			name:      "invalid entity name",
			ifaceName: "ProductReceiver",
			entity:    "product",
			operation: "ProcessProduct",
			wantError: true,
		},
		{
			name:      "invalid operation name",
			ifaceName: "ProductReceiver",
			entity:    "Product",
			operation: "processProduct",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Fatalf("failed to create template engine: %v", err)
			}

			gen := NewInterfaceGenerator(tt.ifaceName, tt.entity, tt.operation, ".", templates)
			err = gen.Validate()

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestInterfaceGenerator_Generate(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create template engine: %v", err)
	}

	tempDir := t.TempDir()
	gen := NewInterfaceGenerator("ProductReceiver", "Product", "ProcessProduct", tempDir, templates)
	if err := gen.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	files, err := gen.Generate()
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	expectedPath := filepath.Join(tempDir, "internal", "domain", "productreceiver.go")
	if files[0].Path != expectedPath {
		t.Errorf("path = %s, want %s", files[0].Path, expectedPath)
	}

	required := []string{
		"package domain",
		"context",
		"type IProductReceiver interface",
		"ProcessProduct(ctx context.Context, product *Product) error",
	}
	for _, check := range required {
		if !strings.Contains(files[0].Content, check) {
			t.Errorf("content missing %q:\n%s", check, files[0].Content)
		}
	}
}

func TestInterfaceGenerator_ComponentType(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create template engine: %v", err)
	}

	gen := NewInterfaceGenerator("ProductReceiver", "Product", "ProcessProduct", ".", templates)
	if gen.ComponentType() != "interface" {
		t.Errorf("ComponentType() = %q, want interface", gen.ComponentType())
	}
}
