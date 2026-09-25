package manager

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

// MainFileManager handles modifications to cmd/main.go
type MainFileManager struct {
	projectDir string
}

// NewMainFileManager creates a new main file manager
func NewMainFileManager(projectDir string) *MainFileManager {
	return &MainFileManager{
		projectDir: projectDir,
	}
}

// AddImport appends importPath to cmd/main.go's import declaration.
// No-op if the import is already present.
func (m *MainFileManager) AddImport(importPath string) error {
	return m.modify(func(fset *token.FileSet, node *ast.File) error {
		if m.hasImport(node, importPath) {
			return errNoChange
		}
		return m.addImport(node, importPath)
	})
}

// AddFxModule inserts moduleExpr (a Go expression such as "usecase.Module")
// into fx.New(...) in cmd/main.go. No-op if an argument with the same
// formatted source is already present. Placed before the first fx.Invoke
// argument when one exists, otherwise appended at the end.
func (m *MainFileManager) AddFxModule(moduleExpr string) error {
	expr, err := parser.ParseExpr(moduleExpr)
	if err != nil {
		return fmt.Errorf("failed to parse module expression %q: %w", moduleExpr, err)
	}

	return m.modify(func(fset *token.FileSet, node *ast.File) error {
		return m.addModuleToFxNew(fset, node, expr)
	})
}

// errNoChange signals that the AST mutation was a no-op and the file
// should not be rewritten.
var errNoChange = fmt.Errorf("no change")

// modify parses cmd/main.go, applies mutate, and writes the file back
// unless mutate returns errNoChange.
func (m *MainFileManager) modify(mutate func(*token.FileSet, *ast.File) error) error {
	mainPath := filepath.Join(m.projectDir, "cmd", "main.go")

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, mainPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse main.go: %w", err)
	}

	if err := mutate(fset, node); err != nil {
		if err == errNoChange {
			return nil
		}
		return err
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return fmt.Errorf("failed to format modified main.go: %w", err)
	}

	if err := os.WriteFile(mainPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	return nil
}

// hasImport checks if an import path already exists.
func (m *MainFileManager) hasImport(node *ast.File, importPath string) bool {
	for _, imp := range node.Imports {
		if imp.Path.Value == `"`+importPath+`"` {
			return true
		}
	}
	return false
}

// addImport appends an import spec to the file's import declaration.
func (m *MainFileManager) addImport(node *ast.File, importPath string) error {
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		genDecl.Specs = append(genDecl.Specs, &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: `"` + importPath + `"`,
			},
		})
		return nil
	}

	return fmt.Errorf("import declaration not found in main.go")
}

// addModuleToFxNew inserts moduleExpr into the fx.New(...) argument list.
// Insertion is idempotent (compared by formatted source) and positions the
// new argument before the first fx.Invoke(...) argument.
func (m *MainFileManager) addModuleToFxNew(fset *token.FileSet, node *ast.File, moduleExpr ast.Expr) error {
	target, err := formatExpr(fset, moduleExpr)
	if err != nil {
		return err
	}

	var mutateErr error
	var handled bool

	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "main" {
			continue
		}

		ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := selExpr.X.(*ast.Ident)
			if !ok || ident.Name != "fx" || selExpr.Sel.Name != "New" {
				return true
			}

			handled = true

			for _, arg := range callExpr.Args {
				existing, err := formatExpr(fset, arg)
				if err != nil {
					mutateErr = err
					return false
				}
				if existing == target {
					return false
				}
			}

			insertPos := len(callExpr.Args)
			for i, arg := range callExpr.Args {
				if m.isFxInvoke(arg) {
					insertPos = i
					break
				}
			}

			callExpr.Args = append(callExpr.Args[:insertPos],
				append([]ast.Expr{moduleExpr}, callExpr.Args[insertPos:]...)...)

			return false
		})

		if mutateErr != nil {
			return mutateErr
		}
		if handled {
			return nil
		}
	}

	return fmt.Errorf("fx.New call not found in main.go")
}

// isFxInvoke checks if an expression is fx.Invoke(...).
func (m *MainFileManager) isFxInvoke(expr ast.Expr) bool {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := selExpr.X.(*ast.Ident)
	return ok && ident.Name == "fx" && selExpr.Sel.Name == "Invoke"
}

// formatExpr renders expr as Go source using fset for position info.
func formatExpr(fset *token.FileSet, expr ast.Expr) (string, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, expr); err != nil {
		return "", fmt.Errorf("failed to format expression: %w", err)
	}
	return buf.String(), nil
}
