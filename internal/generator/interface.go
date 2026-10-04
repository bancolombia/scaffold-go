package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/bancolombia/scaffold-go/internal/validator"
)

type InterfaceGenerator struct {
	Name       string
	Entity     string
	Operation  string
	ProjectDir string
	Templates  *template.TemplateEngine
	Validator  *validator.Validator
}

func NewInterfaceGenerator(name, entity, operation, projectDir string, templates *template.TemplateEngine) *InterfaceGenerator {
	return &InterfaceGenerator{
		Name:       name,
		Entity:     entity,
		Operation:  operation,
		ProjectDir: projectDir,
		Templates:  templates,
		Validator:  validator.New(),
	}
}

func (g *InterfaceGenerator) Validate() error {
	if err := g.Validator.ValidateName(g.Name); err != nil {
		return fmt.Errorf("invalid interface name: %w", err)
	}
	if hasInterfacePrefix(g.Name) {
		return fmt.Errorf("invalid interface name: %q already includes the I prefix; pass the name without it (for example ProductReceiver)", g.Name)
	}
	if err := g.Validator.ValidateName(g.Entity); err != nil {
		return fmt.Errorf("invalid entity name: %w", err)
	}
	if err := g.Validator.ValidateName(g.Operation); err != nil {
		return fmt.Errorf("invalid operation name: %w", err)
	}
	return nil
}

func (g *InterfaceGenerator) Generate() ([]GeneratedFile, error) {
	data := template.DomainInterfaceTemplateData{
		PackageName:   "domain",
		InterfaceName: "I" + g.Name,
		EntityName:    g.Entity,
		OperationName: g.Operation,
		ParamName:     camelCase(g.Entity),
	}

	content, err := g.Templates.Execute("domain/interface/interface.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate interface: %w", err)
	}

	path := filepath.Join(g.ProjectDir, "internal", "domain", strings.ToLower(g.Name)+".go")
	return []GeneratedFile{
		{Path: path, Content: content},
	}, nil
}

func (g *InterfaceGenerator) ComponentType() string {
	return "interface"
}

func hasInterfacePrefix(name string) bool {
	runes := []rune(name)
	if len(runes) < 2 || runes[0] != 'I' {
		return false
	}
	return unicode.IsUpper(runes[1])
}

func camelCase(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
