package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gabheadz/scaffold-go/internal/template"
	"github.com/gabheadz/scaffold-go/internal/validator"
)

type CacheGenerator struct {
	Name       string
	ProjectDir string
	Templates  *template.TemplateEngine
	Validator  *validator.Validator
}

func NewCacheGenerator(name, projectDir string, templates *template.TemplateEngine) *CacheGenerator {
	return &CacheGenerator{
		Name:       name,
		ProjectDir: projectDir,
		Templates:  templates,
		Validator:  validator.New(),
	}
}

func (g *CacheGenerator) Validate() error {
	if err := g.Validator.ValidateName(g.Name); err != nil {
		return fmt.Errorf("invalid cache name: %w", err)
	}
	return nil
}

func (g *CacheGenerator) Generate() ([]GeneratedFile, error) {
	data := template.CacheTemplateData{
		PackageName: "cache",
		CacheName:   g.Name,
		ModulePath:  "",
	}

	cacheContent, err := g.Templates.Execute("da/cache/cache_adapter.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cache: %w", err)
	}

	moduleContent, err := g.Templates.Execute("da/cache/cache_module.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate module: %w", err)
	}

	cachePath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "cache", strings.ToLower(g.Name)+"_cache.go")
	modulePath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "cache", "module.go")

	return []GeneratedFile{
		{Path: cachePath, Content: cacheContent},
		{Path: modulePath, Content: moduleContent},
	}, nil
}

func (g *CacheGenerator) ComponentType() string {
	return "cache"
}
