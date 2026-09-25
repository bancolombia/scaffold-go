package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gabheadz/scaffold-go/internal/template"
	"github.com/gabheadz/scaffold-go/internal/validator"
)

type ApiClientGenerator struct {
	Name       string
	BaseURL    string
	ProjectDir string
	Templates  *template.TemplateEngine
	Validator  *validator.Validator
}

func NewApiClientGenerator(name, baseURL, projectDir string, templates *template.TemplateEngine) *ApiClientGenerator {
	return &ApiClientGenerator{
		Name:       name,
		BaseURL:    baseURL,
		ProjectDir: projectDir,
		Templates:  templates,
		Validator:  validator.New(),
	}
}

func (g *ApiClientGenerator) Validate() error {
	if err := g.Validator.ValidateName(g.Name); err != nil {
		return fmt.Errorf("invalid API client name: %w", err)
	}
	return nil
}

func (g *ApiClientGenerator) Generate() ([]GeneratedFile, error) {
	data := template.ApiClientTemplateData{
		PackageName: "api",
		ClientName:  g.Name,
		BaseURL:     g.BaseURL,
		ModulePath:  "",
	}

	clientContent, err := g.Templates.Execute("da/api/api_client.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate API client: %w", err)
	}

	moduleContent, err := g.Templates.Execute("da/api/api_module.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate module: %w", err)
	}

	clientPath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "api", strings.ToLower(g.Name)+"_client.go")
	modulePath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "api", "module.go")

	return []GeneratedFile{
		{Path: clientPath, Content: clientContent},
		{Path: modulePath, Content: moduleContent},
	}, nil
}

func (g *ApiClientGenerator) ComponentType() string {
	return "api-client"
}
