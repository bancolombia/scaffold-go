package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
)

func TestUseCaseGenerator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ucName  string
		wantErr bool
	}{
		{
			name:    "valid use case name",
			ucName:  "CreateCustomer",
			wantErr: false,
		},
		{
			name:    "valid use case name with multiple words",
			ucName:  "CreateCustomerOrder",
			wantErr: false,
		},
		{
			name:    "invalid use case name with special characters",
			ucName:  "Create-Customer",
			wantErr: true,
		},
		{
			name:    "invalid use case name starting with number",
			ucName:  "1CreateCustomer",
			wantErr: true,
		},
		{
			name:    "empty use case name",
			ucName:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				t.Fatalf("Failed to create template engine: %v", err)
			}

			gen := NewUseCaseGenerator(tt.ucName, []string{}, tmpDir, tmpl)
			err = gen.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUseCaseGenerator_Generate(t *testing.T) {
	tests := []struct {
		name         string
		ucName       string
		dependencies []string
		wantErr      bool
		wantContains []string
	}{
		{
			name:    "generate use case with no dependencies",
			ucName:  "CreateCustomer",
			wantErr: false,
			wantContains: []string{
				"type CreateCustomerUseCase interface",
				"type createcustomerImpl struct",
				"func NewCreateCustomerUseCase(logger *zap.Logger)",
				"func (uc *createcustomerImpl) Execute(ctx context.Context, input any) (any, error)",
			},
		},
		{
			name:   "generate use case with single dependency",
			ucName: "CreateCustomer",
			dependencies: []string{
				"customerRepo:CustomerRepository",
			},
			wantErr: false,
			wantContains: []string{
				"type CreateCustomerUseCase interface",
				"customerRepo domain.CustomerRepository",
				"func NewCreateCustomerUseCase(customerRepo domain.CustomerRepository, logger *zap.Logger)",
			},
		},
		{
			name:   "generate use case with multiple dependencies",
			ucName: "CreateOrder",
			dependencies: []string{
				"orderRepo:OrderRepository",
				"customerRepo:CustomerRepository",
			},
			wantErr: false,
			wantContains: []string{
				"type CreateOrderUseCase interface",
				"orderRepo domain.OrderRepository",
				"customerRepo domain.CustomerRepository",
				"func NewCreateOrderUseCase(orderRepo domain.OrderRepository, customerRepo domain.CustomerRepository, logger *zap.Logger)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				t.Fatalf("Failed to create template engine: %v", err)
			}

			gen := NewUseCaseGenerator(tt.ucName, tt.dependencies, tmpDir, tmpl)
			files, err := gen.Generate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify three files are generated (usecase + module + test)
			if len(files) != 3 {
				t.Errorf("Generate() returned %d files, want 3", len(files))
				return
			}

			// Verify file paths
			expectedPath := filepath.Join(tmpDir, "internal", "usecase", strings.ToLower(tt.ucName)+"_usecase.go")
			expectedModulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			expectedTestPath := filepath.Join(tmpDir, "internal", "usecase", strings.ToLower(tt.ucName)+"_usecase_test.go")
			if files[0].Path != expectedPath {
				t.Errorf("Generate() file path = %v, want %v", files[0].Path, expectedPath)
			}
			if files[1].Path != expectedTestPath {
				t.Errorf("Generate() test file path = %v, want %v", files[1].Path, expectedTestPath)
			}
			if files[2].Path != expectedModulePath {
				t.Errorf("Generate() module file path = %v, want %v", files[2].Path, expectedModulePath)
			}

			// Verify content contains expected strings
			content := files[0].Content
			for _, want := range tt.wantContains {
				if !strings.Contains(content, want) {
					t.Errorf("Generate() content does not contain %q\nGot:\n%s", want, content)
				}
			}

			// Verify it's valid Go code (basic check)
			if !strings.Contains(content, "package usecase") {
				t.Errorf("Generate() content does not contain package declaration")
			}
		})
	}
}

func TestUseCaseGenerator_ComponentType(t *testing.T) {
	tmpDir := t.TempDir()
	tmpl, err := template.NewTemplateEngine()
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	gen := NewUseCaseGenerator("CreateCustomer", []string{}, tmpDir, tmpl)
	if got := gen.ComponentType(); got != "usecase" {
		t.Errorf("ComponentType() = %v, want %v", got, "usecase")
	}
}

func TestUseCaseGenerator_ParseDependencies(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []string
		want         int
		wantNames    []string
		wantTypes    []string
	}{
		{
			name:         "empty dependencies",
			dependencies: []string{},
			want:         0,
		},
		{
			name: "single dependency",
			dependencies: []string{
				"customerRepo:CustomerRepository",
			},
			want:      1,
			wantNames: []string{"customerRepo"},
			wantTypes: []string{"CustomerRepository"},
		},
		{
			name: "multiple dependencies",
			dependencies: []string{
				"orderRepo:OrderRepository",
				"customerRepo:CustomerRepository",
			},
			want:      2,
			wantNames: []string{"orderRepo", "customerRepo"},
			wantTypes: []string{"OrderRepository", "CustomerRepository"},
		},
		{
			name: "invalid format ignored",
			dependencies: []string{
				"orderRepo:OrderRepository",
				"invalid",
				"customerRepo:CustomerRepository",
			},
			want:      2,
			wantNames: []string{"orderRepo", "customerRepo"},
			wantTypes: []string{"OrderRepository", "CustomerRepository"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				t.Fatalf("Failed to create template engine: %v", err)
			}

			gen := NewUseCaseGenerator("TestUseCase", tt.dependencies, tmpDir, tmpl)
			deps := gen.parseDependencies()

			if len(deps) != tt.want {
				t.Errorf("parseDependencies() returned %d dependencies, want %d", len(deps), tt.want)
			}

			for i, dep := range deps {
				if i < len(tt.wantNames) && dep.Name != tt.wantNames[i] {
					t.Errorf("parseDependencies()[%d].Name = %v, want %v", i, dep.Name, tt.wantNames[i])
				}
				if i < len(tt.wantTypes) && dep.Type != tt.wantTypes[i] {
					t.Errorf("parseDependencies()[%d].Type = %v, want %v", i, dep.Type, tt.wantTypes[i])
				}
			}
		})
	}
}

func TestUseCaseGenerator_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	tmpl, err := template.NewTemplateEngine()
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// Create the directory structure
	usecaseDir := filepath.Join(tmpDir, "internal", "usecase")
	if err := os.MkdirAll(usecaseDir, 0755); err != nil {
		t.Fatalf("Failed to create usecase directory: %v", err)
	}

	// Generate use case
	gen := NewUseCaseGenerator("CreateCustomer", []string{"customerRepo:CustomerRepository"}, tmpDir, tmpl)

	if err := gen.Validate(); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	files, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Write the files
	if len(files) != 3 {
		t.Fatalf("Expected 3 files, got %d", len(files))
	}

	for _, file := range files {
		if err := os.WriteFile(file.Path, []byte(file.Content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	filePath := files[0].Path

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Generated file does not exist: %s", filePath)
	}

	// Read and verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	expectedStrings := []string{
		"package usecase",
		"type CreateCustomerUseCase interface",
		"Execute(ctx context.Context, input any) (any, error)",
		"customerRepo domain.CustomerRepository",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Generated file does not contain %q", expected)
		}
	}
}
