package manager

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

// GenericModuleManager handles creation and updates to fx modules
type GenericModuleManager struct {
	projectDir string
}

// NewUsecaseModuleManager creates a new usecase module manager
func NewGenericModuleManager(projectDir string) *GenericModuleManager {
	return &GenericModuleManager{
		projectDir: projectDir,
	}
}

// ModuleExists checks if the usecase module file exists
func (m *GenericModuleManager) ModuleExists(modulePath string) bool {
	_, err := os.Stat(modulePath)
	return err == nil
}

// BuildCreateContent returns the content for a new usecase module file with the
// given use case provider, without touching the filesystem.
func (m *GenericModuleManager) BuildCreateContent(initialContent string, placeholder string) string {
	return fmt.Sprintf(initialContent, placeholder)
}

// CreateModule creates a new usecase module file with the given use case provider
func (m *GenericModuleManager) CreateModule(moduleDir string, modulePath string, initialContent string, placeholder string) error {
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}

	content := m.BuildCreateContent(initialContent, placeholder)

	if err := os.WriteFile(modulePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create module.go: %w", err)
	}

	return nil
}

// BuildAddProviderContent reads the existing module.go, adds a provider for the
// given use case via AST manipulation, and returns the updated content without
// writing it. If the provider already exists, the original content is returned
// unchanged (idempotent).
func (m *GenericModuleManager) BuildAddProviderContent(modulePath string, providerName string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, modulePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse module.go: %w", err)
	}

	// if the use case provider already exists, return the original content
	if m.hasProvider(node, providerName) {
		original, readErr := os.ReadFile(modulePath)
		if readErr != nil {
			return "", fmt.Errorf("failed to read module.go: %w", readErr)
		}
		return string(original), nil
	}

	// add the new provider to the fx.Provide call
	if err := m.addProviderToFxProvide(node, providerName); err != nil {
		return "", fmt.Errorf("failed to add provider: %w", err)
	}

	// format the modified AST back into Go source code
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "", fmt.Errorf("failed to format modified module.go: %w", err)
	}

	return buf.String(), nil
}

// AddProvider adds a new provider to the existing module file
// Returns nil if provider already exists (idempotent operation)
func (m *GenericModuleManager) AddProvider(modulePath string, providerName string) error {

	content, err := m.BuildAddProviderContent(modulePath, providerName)
	if err != nil {
		return err
	}

	if err := os.WriteFile(modulePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", modulePath, err)
	}

	return nil
}

// hasProvider checks if a provider function already exists in fx.Provide
func (m *GenericModuleManager) hasProvider(node *ast.File, providerName string) bool {
	found := false

	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if this is fx.Provide
		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok || ident.Name != "fx" || selExpr.Sel.Name != "Provide" {
			return true
		}

		// Check arguments for provider name
		for _, arg := range callExpr.Args {
			if argIdent, ok := arg.(*ast.Ident); ok && argIdent.Name == providerName {
				found = true
				return false
			}
		}

		return true
	})

	return found
}

// addProviderToFxProvide adds a provider function to fx.Provide call
func (m *GenericModuleManager) addProviderToFxProvide(node *ast.File, providerName string) error {
	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if this is fx.Provide
		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok || ident.Name != "fx" || selExpr.Sel.Name != "Provide" {
			return true
		}

		// Add provider at the end
		newProvider := &ast.Ident{Name: providerName}
		callExpr.Args = append(callExpr.Args, newProvider)

		return false
	})

	return nil
}
