package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gabheadz/scaffold-go/internal/manager"
	"github.com/gabheadz/scaffold-go/internal/template"
	"github.com/gabheadz/scaffold-go/internal/validator"
)

type UseCaseGenerator struct {
	Name         string
	Dependencies []string
	ProjectDir   string
	Templates    *template.TemplateEngine
	Validator    *validator.Validator
}

func NewUseCaseGenerator(name string, dependencies []string, projectDir string, templates *template.TemplateEngine) *UseCaseGenerator {
	return &UseCaseGenerator{
		Name:         name,
		Dependencies: dependencies,
		ProjectDir:   projectDir,
		Templates:    templates,
		Validator:    validator.New(),
	}
}

func (g *UseCaseGenerator) Validate() error {
	if err := g.Validator.ValidateName(g.Name); err != nil {
		return fmt.Errorf("invalid use case name: %w", err)
	}
	return nil
}

func (g *UseCaseGenerator) Generate() ([]GeneratedFile, error) {

	files := []GeneratedFile{}

	deps := g.parseDependencies()

	// Get module path from go.mod
	mp, err := g.Validator.GetModulePath(g.ProjectDir)
	if err != nil {
		// If we can't get module path, use empty string (tests might not have go.mod)
		mp = ""
	}

	data := template.UseCaseTemplateData{
		PackageName:  "usecase",
		UseCaseName:  g.Name,
		Dependencies: deps,
		InputType:    "any",
		OutputType:   "any",
		ModulePath:   mp,
		HasLogger:    true, // Always include logger for structured logging
	}

	// -------------------------------------
	// ----- Generate Use Case & Tests -----
	// -------------------------------------
	content, err := g.Templates.Execute("domain/usecase/usecase.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate use case: %w", err)
	}
	path := filepath.Join(g.ProjectDir, "internal", "usecase", strings.ToLower(g.Name)+"_usecase.go")
	files = append(files, GeneratedFile{Path: path, Content: content})

	testContent, err := g.Templates.Execute("domain/usecase/usecase_test.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate use case test: %w", err)
	}
	testPath := filepath.Join(g.ProjectDir, "internal", "usecase", strings.ToLower(g.Name)+"_usecase_test.go")
	files = append(files, GeneratedFile{Path: testPath, Content: testContent})

	// --------------------------------------------
	// ----- Generate or modify Usecase Module -----
	// --------------------------------------------
	modulePath := filepath.Join(g.ProjectDir, "internal", "usecase", "module.go")
	moduleManager := manager.NewGenericModuleManager(g.ProjectDir)

	var moduleContent string
	if moduleManager.ModuleExists(modulePath) {
		updated, err := moduleManager.BuildAddProviderContent(modulePath, "New"+g.Name+"UseCase")
		if err != nil {
			return nil, fmt.Errorf("failed to add provider to usecase module for use case '%s': %w", g.Name, err)
		}
		moduleContent = updated
	} else {
		moduleContent = moduleManager.BuildCreateContent(provideStarterContentForModule(), "New"+g.Name+"UseCase")
	}
	files = append(files, GeneratedFile{Path: modulePath, Content: moduleContent})

	return files, nil
}

func (g *UseCaseGenerator) ComponentType() string {
	return "usecase"
}

// extractEntityType extracts the entity name from a repository type.
// For example, "IUserRepository" returns "User", "OrderRepository" returns "Order".
// Returns empty string for non-repository types or edge cases like "Repository" alone.
func extractEntityType(repoType string) string {
	const repositorySuffix = "Repository"
	const repositoryPrefix = "I"

	// Check if type ends with "Repository"
	if !strings.HasSuffix(repoType, repositorySuffix) {
		return ""
	}

	// Extract the prefix (entity name)
	entityName := strings.TrimPrefix(strings.TrimSuffix(repoType, repositorySuffix), repositoryPrefix)

	// Handle edge cases: empty names or "Repository" alone
	if entityName == "" {
		return ""
	}

	return entityName
}

func (g *UseCaseGenerator) parseDependencies() []template.Dependency {
	deps := make([]template.Dependency, 0, len(g.Dependencies))
	for _, d := range g.Dependencies {
		parts := strings.Split(d, ":")
		if len(parts) == 2 {
			depType := strings.TrimSpace(parts[1])
			deps = append(deps, template.Dependency{
				Name:       strings.TrimSpace(parts[0]),
				Type:       depType,
				EntityType: extractEntityType(depType),
			})
		}
	}
	return deps
}

func provideStarterContentForModule() string {
	return `package usecase
import "go.uber.org/fx"

// Module provides use cases dependencies as an fx Module
var Module = fx.Module("usecase",
	fx.Provide(
		%s,
	),
)
`
}
