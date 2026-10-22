package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/bancolombia/scaffold-go/internal/validator"
)

type EntityGenerator struct {
	Name       string
	Fields     []string
	ProjectDir string
	Templates  *template.TemplateEngine
	Validator  *validator.Validator
}

func NewEntityGenerator(name string, fields []string, projectDir string, templates *template.TemplateEngine) *EntityGenerator {
	return &EntityGenerator{
		Name:       name,
		Fields:     fields,
		ProjectDir: projectDir,
		Templates:  templates,
		Validator:  validator.New(),
	}
}

func (g *EntityGenerator) Validate() error {
	if err := g.Validator.ValidateName(g.Name); err != nil {
		return fmt.Errorf("invalid entity name: %w", err)
	}
	if err := g.Validator.ValidateFields(g.Fields); err != nil {
		return fmt.Errorf("invalid fields: %w", err)
	}
	return nil
}

func (g *EntityGenerator) Generate() ([]GeneratedFile, error) {
	fields := g.parseFields()

	// Check if any field is of type time.Time
	hasTimeFields := false
	for _, f := range fields {
		if strings.Contains(f.Type, "time.Time") {
			hasTimeFields = true
			break
		}
	}

	entityData := template.EntityTemplateData{
		PackageName:    "domain",
		EntityName:     g.Name,
		Fields:         fields,
		HasTimeFields:  hasTimeFields,
		HasValidation:  true,
		RepositoryName: "I" + g.Name + "Repository",
		ModulePath:     "", // Module path is not needed for entity/repository templates
	}

	entityContent, err := g.Templates.Execute("domain/entity/entity.go.tmpl", entityData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entity: %w", err)
	}

	repoContent, err := g.Templates.Execute("domain/entity/repository.go.tmpl", entityData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate repository: %w", err)
	}

	testContent, err := g.Templates.Execute("domain/entity/entity_test.go.tmpl", entityData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entity test: %w", err)
	}

	entityPath := filepath.Join(g.ProjectDir, "internal", "domain", strings.ToLower(g.Name)+".go")
	repoPath := filepath.Join(g.ProjectDir, "internal", "domain", strings.ToLower(g.Name)+"_repository.go")
	testPath := filepath.Join(g.ProjectDir, "internal", "domain", strings.ToLower(g.Name)+"_test.go")

	return []GeneratedFile{
		{Path: entityPath, Content: entityContent},
		{Path: repoPath, Content: repoContent},
		{Path: testPath, Content: testContent},
	}, nil
}

func (g *EntityGenerator) ComponentType() string {
	return "entity"
}

func (g *EntityGenerator) parseFields() []template.Field {
	fields := make([]template.Field, 0, len(g.Fields))
	for _, f := range g.Fields {
		parts := strings.Split(f, ":")
		fields = append(fields, template.Field{
			Name: strings.TrimSpace(parts[0]),
			Type: strings.TrimSpace(parts[1]),
		})
	}
	return fields
}
