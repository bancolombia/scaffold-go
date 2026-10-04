package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func genValidInterfaceName() gopter.Gen {
	// Avoid names that already include the I prefix (I + uppercase).
	return gen.RegexMatch(`^[A-HJ-Z][a-zA-Z0-9]{0,20}`)
}

// Feature: go-microservice-scaffolding-cli, Property: domain interface generation
func TestProperty_InterfaceGenerationDeclaresPort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping property-based test in short mode")
	}

	parameters := GetTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("generated interface uses I prefix, pointer entity, and lowercase path", prop.ForAll(
		func(ifaceName, entityName, operationName string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			gen := NewInterfaceGenerator(ifaceName, entityName, operationName, tmpDir, templates)
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}
			if len(files) != 1 {
				t.Logf("expected 1 file, got %d", len(files))
				return false
			}

			expectedPath := filepath.Join(tmpDir, "internal", "domain", strings.ToLower(ifaceName)+".go")
			if files[0].Path != expectedPath {
				t.Logf("path = %s, want %s", files[0].Path, expectedPath)
				return false
			}

			content := files[0].Content
			checks := []string{
				"type I" + ifaceName + " interface",
				operationName + "(ctx context.Context, " + camelCase(entityName) + " *" + entityName + ") error",
			}
			for _, check := range checks {
				if !strings.Contains(content, check) {
					t.Logf("content missing %q", check)
					return false
				}
			}
			return true
		},
		genValidInterfaceName(),
		genValidEntityName(),
		genValidInterfaceName(),
	))

	properties.TestingRun(t)
}
