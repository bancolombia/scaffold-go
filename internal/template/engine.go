package template

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
)

//go:embed all:templates
var templatesFS embed.FS

type TemplateEngine struct {
	templates *template.Template
}

func NewTemplateEngine() (*TemplateEngine, error) {
	tmpl := template.New("").Funcs(getFuncMap())

	walkErr := fs.WalkDir(templatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		content, readErr := templatesFS.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("failed to read template %s: %w", path, readErr)
		}
		// Use path relative to "templates/" as the template name so that
		// generators can reference templates by their layered path
		// (e.g. "ep/http/http_handler.go.tmpl").
		name := strings.TrimPrefix(path, "templates/")
		if _, parseErr := tmpl.New(name).Parse(string(content)); parseErr != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, parseErr)
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("failed to load templates: %w", walkErr)
	}

	return &TemplateEngine{templates: tmpl}, nil
}

func (te *TemplateEngine) Execute(templateName string, data interface{}) (string, error) {
	tmpl := te.templates.Lookup(templateName)
	if tmpl == nil {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"toSnakeCase":  toSnakeCase,
		"toCamelCase":  toCamelCase,
		"toKebabCase":  toKebabCase,
		"pluralize":    pluralize,
		"toLower":      strings.ToLower,
		"toUpper":      strings.ToUpper,
		"zeroValue":    zeroValue,
		"testValue":    testValue,
		"toTableName":  toTableName,
		"toColumnName": toColumnName,
		"add":          func(a, b int) int { return a + b },
		"isLast":       func(index, length int) bool { return index == length-1 },
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func toKebabCase(s string) string {
	return strings.ReplaceAll(toSnakeCase(s), "_", "-")
}

func pluralize(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") {
		return s + "es"
	}
	return s + "s"
}

func zeroValue(typ string) string {
	switch typ {
	case "string":
		return `""`
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "bool":
		return "false"
	default:
		return "nil"
	}
}

func testValue(typ string) string {
	switch typ {
	case "string":
		return `"test-value"`
	case "int", "int32", "int64":
		return "123"
	case "uint", "uint32", "uint64":
		return "123"
	case "float32", "float64":
		return "123.45"
	case "bool":
		return "true"
	case "time.Time":
		return "time.Now()"
	default:
		return "nil"
	}
}

func toTableName(s string) string {
	return pluralize(toSnakeCase(s))
}

func toColumnName(s string) string {
	return toSnakeCase(s)
}
