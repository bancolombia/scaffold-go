package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
)

func TestEntityGenerator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		entName   string
		fields    []string
		wantError bool
	}{
		{
			name:      "valid entity with fields",
			entName:   "Customer",
			fields:    []string{"Name:string", "Email:string", "Age:int"},
			wantError: false,
		},
		{
			name:      "valid entity without fields",
			entName:   "Product",
			fields:    []string{},
			wantError: false,
		},
		{
			name:      "invalid entity name with special chars",
			entName:   "Invalid-Name",
			fields:    []string{"Name:string"},
			wantError: true,
		},
		{
			name:      "invalid entity name starting with number",
			entName:   "123Entity",
			fields:    []string{"Name:string"},
			wantError: true,
		},
		{
			name:      "empty entity name",
			entName:   "",
			fields:    []string{"Name:string"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Fatalf("failed to create template engine: %v", err)
			}

			gen := NewEntityGenerator(tt.entName, tt.fields, ".", templates)
			err = gen.Validate()

			if tt.wantError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestEntityGenerator_Generate(t *testing.T) {
	tests := []struct {
		name           string
		entName        string
		fields         []string
		wantEntityFile bool
		wantRepoFile   bool
		wantTestFile   bool
	}{
		{
			name:           "generate entity with fields",
			entName:        "Customer",
			fields:         []string{"Name:string", "Email:string", "Age:int"},
			wantEntityFile: true,
			wantRepoFile:   true,
			wantTestFile:   true,
		},
		{
			name:           "generate entity without fields",
			entName:        "Product",
			fields:         []string{},
			wantEntityFile: true,
			wantRepoFile:   true,
			wantTestFile:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "entity-gen-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Fatalf("failed to create template engine: %v", err)
			}

			// Create generator
			gen := NewEntityGenerator(tt.entName, tt.fields, tempDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}

			// Check number of files
			expectedFiles := 0
			if tt.wantEntityFile {
				expectedFiles++
			}
			if tt.wantRepoFile {
				expectedFiles++
			}
			if tt.wantTestFile {
				expectedFiles++
			}

			if len(files) != expectedFiles {
				t.Errorf("expected %d files, got %d", expectedFiles, len(files))
			}

			// Check file paths and content
			for _, file := range files {
				var requiredContent []string

				if strings.Contains(file.Path, "_test.go") {
					if !tt.wantTestFile {
						t.Errorf("unexpected test file generated")
					}
					// Check test file path
					expectedPath := filepath.Join(tempDir, "internal", "domain", strings.ToLower(tt.entName)+"_test.go")
					if file.Path != expectedPath {
						t.Errorf("expected test path %s, got %s", expectedPath, file.Path)
					}
					// Test-specific content checks
					requiredContent = []string{"func TestNew" + tt.entName, "func Test" + tt.entName + "_Validate", "setupTest"}
				} else if strings.Contains(file.Path, "_repository.go") {
					if !tt.wantRepoFile {
						t.Errorf("unexpected repository file generated")
					}
					// Check repository file path
					expectedPath := filepath.Join(tempDir, "internal", "domain", strings.ToLower(tt.entName)+"_repository.go")
					if file.Path != expectedPath {
						t.Errorf("expected repository path %s, got %s", expectedPath, file.Path)
					}
					// Repository-specific content checks
					requiredContent = []string{tt.entName + "Repository", "Create", "GetByID", "Update", "Delete", "List"}
				} else {
					if !tt.wantEntityFile {
						t.Errorf("unexpected entity file generated")
					}
					// Check entity file path
					expectedPath := filepath.Join(tempDir, "internal", "domain", strings.ToLower(tt.entName)+".go")
					if file.Path != expectedPath {
						t.Errorf("expected entity path %s, got %s", expectedPath, file.Path)
					}
					// Entity-specific content checks
					requiredContent = []string{"type " + tt.entName + " struct"}
					for _, field := range tt.fields {
						parts := strings.Split(field, ":")
						requiredContent = append(requiredContent, parts[0]+" "+parts[1])
					}
				}

				// Check content
				for _, check := range requiredContent {
					if !strings.Contains(file.Content, check) {
						t.Errorf("expected content to contain %q, but it doesn't.\nFile: %s\nContent:\n%s", check, filepath.Base(file.Path), file.Content)
					}
				}
			}
		})
	}
}

func TestEntityGenerator_ComponentType(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create template engine: %v", err)
	}

	gen := NewEntityGenerator("Test", []string{}, ".", templates)
	if gen.ComponentType() != "entity" {
		t.Errorf("expected component type 'entity', got %s", gen.ComponentType())
	}
}

func TestEntityGenerator_ParseFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want   int
	}{
		{
			name:   "parse multiple fields",
			fields: []string{"Name:string", "Email:string", "Age:int"},
			want:   3,
		},
		{
			name:   "parse single field",
			fields: []string{"ID:string"},
			want:   1,
		},
		{
			name:   "parse no fields",
			fields: []string{},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Fatalf("failed to create template engine: %v", err)
			}

			gen := NewEntityGenerator("Test", tt.fields, ".", templates)
			parsed := gen.parseFields()

			if len(parsed) != tt.want {
				t.Errorf("expected %d fields, got %d", tt.want, len(parsed))
			}

			// Verify field structure
			for i, field := range parsed {
				parts := strings.Split(tt.fields[i], ":")
				if field.Name != strings.TrimSpace(parts[0]) {
					t.Errorf("field %d: expected name %s, got %s", i, parts[0], field.Name)
				}
				if field.Type != strings.TrimSpace(parts[1]) {
					t.Errorf("field %d: expected type %s, got %s", i, parts[1], field.Type)
				}
			}
		})
	}
}
