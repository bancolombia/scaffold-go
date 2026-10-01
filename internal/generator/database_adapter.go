package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/bancolombia/scaffold-go/internal/validator"
)

type DatabaseAdapterGenerator struct {
	DatabaseType string
	EntityName   string
	Fields       []string
	ProjectDir   string
	Templates    *template.TemplateEngine
	Validator    *validator.Validator
}

func NewDatabaseAdapterGenerator(dbType, entityName string, fields []string, projectDir string, templates *template.TemplateEngine) *DatabaseAdapterGenerator {
	return &DatabaseAdapterGenerator{
		DatabaseType: dbType,
		EntityName:   entityName,
		Fields:       fields,
		ProjectDir:   projectDir,
		Templates:    templates,
		Validator:    validator.New(),
	}
}

func (g *DatabaseAdapterGenerator) Validate() error {
	if err := g.Validator.ValidateName(g.EntityName); err != nil {
		return fmt.Errorf("invalid entity name: %w", err)
	}
	validTypes := map[string]bool{"postgres": true, "mongodb": true, "dynamodb": true}
	if !validTypes[strings.ToLower(g.DatabaseType)] {
		return fmt.Errorf("invalid database type: %s (must be postgres, mongodb, or dynamodb)", g.DatabaseType)
	}
	if g.ProjectDir != "" {
		entityFile := filepath.Join(g.ProjectDir, "internal", "domain", strings.ToLower(g.EntityName)+".go")
		repoFile := filepath.Join(g.ProjectDir, "internal", "domain", strings.ToLower(g.EntityName)+"_repository.go")
		if _, err := os.Stat(entityFile); os.IsNotExist(err) {
			return fmt.Errorf("entity %q not found at %s — run `scaffold-go generate entity %s ...` first", g.EntityName, entityFile, g.EntityName)
		}
		if _, err := os.Stat(repoFile); os.IsNotExist(err) {
			return fmt.Errorf("repository interface for %q not found at %s — run `scaffold-go generate entity %s ...` first", g.EntityName, repoFile, g.EntityName)
		}
	}
	return nil
}

func (g *DatabaseAdapterGenerator) Generate() ([]GeneratedFile, error) {
	fields := g.parseFields()

	// Get module path from go.mod
	modulePath, err := g.Validator.GetModulePath(g.ProjectDir)
	if err != nil {
		// If we can't get module path, use empty string (tests might not have go.mod)
		modulePath = ""
	}

	data := template.AdapterTemplateData{
		PackageName:   "database",
		InterfaceName: "I" + g.EntityName + "Repository",
		DatabaseType:  strings.ToLower(g.DatabaseType),
		EntityName:    g.EntityName,
		Fields:        fields,
		Methods:       []template.Method{},
		ModulePath:    modulePath,
	}

	generatedFiles := []GeneratedFile{}

	// ---------------------------------
	// --- Generate database adapter ---
	// ---------------------------------
	adapterPath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "database", strings.ToLower(g.DatabaseType)+"_adapter.go")
	if err := g.Validator.FileExists(adapterPath); err == nil {
		// If already exists then ignore re-generation
	} else {
		templateName := fmt.Sprintf("da/database/%s_adapter.go.tmpl", strings.ToLower(g.DatabaseType))
		adapterContent, adapterErr := g.Templates.Execute(templateName, data)
		if adapterErr != nil {
			return nil, fmt.Errorf("failed to generate database adapter: %w", adapterErr)
		}
		generatedFiles = append(generatedFiles, GeneratedFile{Path: adapterPath, Content: adapterContent})
	}

	// -----------------------------------------------------
	// --- Generate database repository for named entity ---
	// -----------------------------------------------------
	repoPath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "database", strings.ToLower(g.EntityName)+"_repository.go")
	repoTemplateName := fmt.Sprintf("da/database/%s_repository.go.tmpl", strings.ToLower(g.DatabaseType))
	repoContent, repoErr := g.Templates.Execute(repoTemplateName, data)
	if repoErr != nil {
		return nil, fmt.Errorf("failed to generate database entity repository: %w", repoErr)
	}
	generatedFiles = append(generatedFiles, GeneratedFile{Path: repoPath, Content: repoContent})

	// --------------------------------------
	// --- Generate database repository test ---
	// --------------------------------------
	testPath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "database", strings.ToLower(g.EntityName)+"_repository_test.go")
	testContent, testErr := g.Templates.Execute(fmt.Sprintf("da/database/%s_repository_test.go.tmpl", strings.ToLower(g.DatabaseType)), data)
	if testErr != nil {
		return nil, fmt.Errorf("failed to generate repository test: %w", testErr)
	}
	generatedFiles = append(generatedFiles, GeneratedFile{Path: testPath, Content: testContent})

	// --------------------------------
	// --- Generate database module ---
	// --------------------------------
	moduleFilePath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "database", "module.go")
	if err := g.Validator.FileExists(moduleFilePath); err == nil {
		// The database module already exists, no need to re-generate it
	} else {
		// create the database module file since it doesn't exist
		moduleContent, moduleErr := g.Templates.Execute("da/database/database_module.go.tmpl", data)
		if moduleErr != nil {
			return nil, fmt.Errorf("failed to generate module: %w", moduleErr)
		}
		generatedFiles = append(generatedFiles, GeneratedFile{Path: moduleFilePath, Content: moduleContent})
	}

	// Ensure the main.go file imports the database module for auto-wiring, if already present it will not duplicate the import
	if err := EnsureMainImport(g.ProjectDir, modulePath+"/internal/drivenadapters/database", ""); err != nil {
		fmt.Printf("⚠ auto-wire import failed: %s\n", err.Error())
	}
	// Ensure the main.go file has the database module wired up, if already wired it will not duplicate it
	if err := EnsureMainModule(g.ProjectDir, "database.Module"); err != nil {
		fmt.Printf("⚠ auto-wire module failed: %s\n", err.Error())
	}

	return generatedFiles, nil
}

func (g *DatabaseAdapterGenerator) ComponentType() string {
	return "database-adapter"
}

func (g *DatabaseAdapterGenerator) parseFields() []template.Field {
	fields := make([]template.Field, 0, len(g.Fields))
	for _, f := range g.Fields {
		parts := strings.Split(f, ":")
		if len(parts) == 2 {
			fields = append(fields, template.Field{
				Name: strings.TrimSpace(parts[0]),
				Type: strings.TrimSpace(parts[1]),
			})
		}
	}
	return fields
}
