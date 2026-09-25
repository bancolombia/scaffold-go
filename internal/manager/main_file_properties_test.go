package manager

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Custom generators for main file property-based testing

// genSingleSegmentModulePath generates single-segment module paths
func genSingleSegmentModulePath() gopter.Gen {
	return gen.RegexMatch(`^[a-z][a-z0-9]{2,15}$`)
}

// genTwoSegmentModulePath generates two-segment module paths
func genTwoSegmentModulePath() gopter.Gen {
	return gen.RegexMatch(`^[a-z][a-z0-9]{2,10}\.[a-z]{2,5}/[a-z][a-z0-9]{2,15}$`)
}

// genThreeSegmentModulePath generates three-segment module paths
func genThreeSegmentModulePath() gopter.Gen {
	return gen.RegexMatch(`^[a-z][a-z0-9]{2,10}\.[a-z]{2,5}/[a-z][a-z0-9]{2,15}/[a-z][a-z0-9]{2,15}$`)
}

// genModulePathWithHyphens generates module paths containing hyphens
func genModulePathWithHyphens() gopter.Gen {
	return gen.OneConstOf(
		"github.com/user/my-project",
		"gitlab.com/org/my-service",
		"example.com/my-app",
		"github.com/my-org/my-microservice",
		"bitbucket.org/team/my-api",
	)
}

// genModulePathWithUnderscores generates module paths containing underscores
func genModulePathWithUnderscores() gopter.Gen {
	return gen.OneConstOf(
		"github.com/user/my_project",
		"gitlab.com/org/my_service",
		"example.com/my_app",
		"github.com/my_org/my_microservice",
		"bitbucket.org/team/my_api",
	)
}

// genValidModulePath generates any valid module path
func genValidModulePath() gopter.Gen {
	return gen.OneGenOf(
		genSingleSegmentModulePath(),
		genTwoSegmentModulePath(),
		genThreeSegmentModulePath(),
		genModulePathWithHyphens(),
		genModulePathWithUnderscores(),
	)
}

// createMainGoFile creates a minimal main.go file with the given module path
func createMainGoFile(projectDir, modulePath string) error {
	cmdDir := filepath.Join(projectDir, "cmd")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}

	content := `package main

import (
	"context"

	"` + modulePath + `/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			zap.NewProduction,
		),
		fx.Invoke(run),
	)

	app.Run()
}

func run(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting application")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping application")
			return nil
		},
	})
}
`
	return os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(content), 0644)
}

// addUsecaseModuleForTest applies the two-call sequence that registers the
// usecase feature in cmd/main.go: it adds the usecase import derived from
// modulePath and inserts usecase.Module into fx.New.
func addUsecaseModuleForTest(mgr *MainFileManager, modulePath string) error {
	if err := mgr.AddImport(modulePath + "/internal/usecase"); err != nil {
		return err
	}
	return mgr.AddFxModule("usecase.Module")
}

// getUsecaseImportFromFile extracts the usecase import path from a parsed main.go
func getUsecaseImportFromFile(mainPath string) (string, bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, mainPath, nil, parser.ParseComments)
	if err != nil {
		return "", false, err
	}

	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		if strings.HasSuffix(importPath, "/internal/usecase") {
			return importPath, true, nil
		}
	}

	return "", false, nil
}

// countImportOccurrences counts how many times a specific import appears in the file
func countImportOccurrences(mainPath, importPath string) (int, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, mainPath, nil, parser.ParseComments)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, imp := range node.Imports {
		if imp.Path.Value == `"`+importPath+`"` {
			count++
		}
	}
	return count, nil
}

// isValidGoFile checks if a file is syntactically valid Go code
func isValidGoFile(filePath string) bool {
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	return err == nil
}

// hasUsecaseModuleInFxNew checks if usecase.Module is present in fx.New arguments
func hasUsecaseModuleInFxNew(mainPath string) (bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, mainPath, nil, parser.ParseComments)
	if err != nil {
		return false, err
	}

	found := false
	ast.Inspect(node, func(n ast.Node) bool {
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
		for _, arg := range callExpr.Args {
			sel, ok := arg.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			id, ok := sel.X.(*ast.Ident)
			if ok && id.Name == "usecase" && sel.Sel.Name == "Module" {
				found = true
				return false
			}
		}
		return true
	})

	return found, nil
}

// Feature: usecase-module-management, Property 3: Usecase Import Addition with Correct Format (usecase.Module also added)
// **Validates: Requirements 2.2, 2.3**
func TestProperty_DomainImportAdditionFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("usecase import is added with format {modulePath}/internal/usecase", prop.ForAll(
		func(modulePath string) bool {
			tmpDir := t.TempDir()

			// Create main.go with the given module path
			if err := createMainGoFile(tmpDir, modulePath); err != nil {
				t.Logf("failed to create main.go: %v", err)
				return false
			}

			// Add domain module
			mgr := NewMainFileManager(tmpDir)
			if err := addUsecaseModuleForTest(mgr, modulePath); err != nil {
				t.Logf("failed to add usecase module: %v", err)
				return false
			}

			// Verify the usecase import was added with correct format
			mainPath := filepath.Join(tmpDir, "cmd", "main.go")
			importPath, found, err := getUsecaseImportFromFile(mainPath)
			if err != nil {
				t.Logf("failed to parse main.go: %v", err)
				return false
			}

			if !found {
				t.Logf("usecase import not found in main.go")
				return false
			}

			expectedImport := modulePath + "/internal/usecase"
			if importPath != expectedImport {
				t.Logf("expected import %s, got %s", expectedImport, importPath)
				return false
			}

			return true
		},
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 3: Usecase Import Addition with Correct Format (Syntactic Validity)
// **Validates: Requirements 2.2, 2.3**
func TestProperty_DomainImportSyntacticallyValid(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("main.go remains syntactically valid after adding usecase import", prop.ForAll(
		func(modulePath string) bool {
			tmpDir := t.TempDir()

			if err := createMainGoFile(tmpDir, modulePath); err != nil {
				t.Logf("failed to create main.go: %v", err)
				return false
			}

			mgr := NewMainFileManager(tmpDir)
			if err := addUsecaseModuleForTest(mgr, modulePath); err != nil {
				t.Logf("failed to add usecase module: %v", err)
				return false
			}

			mainPath := filepath.Join(tmpDir, "cmd", "main.go")
			if !isValidGoFile(mainPath) {
				t.Logf("main.go is not valid Go after adding usecase import")
				return false
			}

			return true
		},
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 3: Usecase Import Addition with Correct Format (Exactly Once)
// **Validates: Requirements 2.2, 2.3**
func TestProperty_DomainImportAppearsExactlyOnce(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("usecase import appears exactly once after addition", prop.ForAll(
		func(modulePath string) bool {
			tmpDir := t.TempDir()

			if err := createMainGoFile(tmpDir, modulePath); err != nil {
				t.Logf("failed to create main.go: %v", err)
				return false
			}

			mgr := NewMainFileManager(tmpDir)
			if err := addUsecaseModuleForTest(mgr, modulePath); err != nil {
				t.Logf("failed to add usecase module: %v", err)
				return false
			}

			mainPath := filepath.Join(tmpDir, "cmd", "main.go")
			expectedImport := modulePath + "/internal/usecase"
			count, err := countImportOccurrences(mainPath, expectedImport)
			if err != nil {
				t.Logf("failed to count imports: %v", err)
				return false
			}

			if count != 1 {
				t.Logf("expected usecase import to appear exactly once, got %d", count)
				return false
			}

			return true
		},
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: domain-module-management, Property 3: Domain Import Addition with Correct Format (With Hyphens)
// **Validates: Requirements 2.2, 2.3**
func TestProperty_UsecaseImportWithHyphenatedPaths(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("usecase import is correctly formatted for paths with hyphens", prop.ForAll(
		func(modulePath string) bool {
			tmpDir := t.TempDir()

			if err := createMainGoFile(tmpDir, modulePath); err != nil {
				t.Logf("failed to create main.go: %v", err)
				return false
			}

			mgr := NewMainFileManager(tmpDir)
			if err := addUsecaseModuleForTest(mgr, modulePath); err != nil {
				t.Logf("failed to add usecase module: %v", err)
				return false
			}

			mainPath := filepath.Join(tmpDir, "cmd", "main.go")
			importPath, found, err := getUsecaseImportFromFile(mainPath)
			if err != nil {
				t.Logf("failed to parse main.go: %v", err)
				return false
			}

			if !found {
				t.Logf("usecase import not found")
				return false
			}

			// Verify the import preserves hyphens from the module path
			if !strings.Contains(importPath, "-") {
				t.Logf("expected import to contain hyphens, got %s", importPath)
				return false
			}

			expectedImport := modulePath + "/internal/usecase"
			if importPath != expectedImport {
				t.Logf("expected import %s, got %s", expectedImport, importPath)
				return false
			}

			return true
		},
		genModulePathWithHyphens(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 3: Usecase Import Addition with Correct Format (domain.Module also added)
// **Validates: Requirements 2.2, 2.3**
func TestProperty_DomainImportAlsoAddsDomainModule(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("adding usecase import also registers domain.Module in fx.New", prop.ForAll(
		func(modulePath string) bool {
			tmpDir := t.TempDir()

			if err := createMainGoFile(tmpDir, modulePath); err != nil {
				t.Logf("failed to create main.go: %v", err)
				return false
			}

			mgr := NewMainFileManager(tmpDir)
			if err := addUsecaseModuleForTest(mgr, modulePath); err != nil {
				t.Logf("failed to add usecase module: %v", err)
				return false
			}

			mainPath := filepath.Join(tmpDir, "cmd", "main.go")
			found, err := hasUsecaseModuleInFxNew(mainPath)
			if err != nil {
				t.Logf("failed to check fx.New: %v", err)
				return false
			}

			if !found {
				t.Logf("usecase.Module not found in fx.New")
				return false
			}

			return true
		},
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// genImportList generates random lists of import paths from standard library and third-party packages
func genImportList() gopter.Gen {
	return gen.SliceOfN(5, gen.OneConstOf(
		"fmt",
		"os",
		"strings",
		"net/http",
		"encoding/json",
		"io",
		"time",
		"sync",
		"errors",
		"github.com/gin-gonic/gin",
		"github.com/spf13/cobra",
		"github.com/spf13/viper",
	)).Map(func(imports []string) []string {
		seen := make(map[string]bool)
		var result []string
		for _, imp := range imports {
			if !seen[imp] {
				seen[imp] = true
				result = append(result, imp)
			}
		}
		return result
	})
}

// createMainGoWithImports creates a main.go with base imports plus extra imports
func createMainGoWithImports(projectDir, modulePath string, extraImports []string) error {
	cmdDir := filepath.Join(projectDir, "cmd")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}

	// Build the import block: base imports + extra imports
	var importBlock strings.Builder
	importBlock.WriteString("import (\n")
	importBlock.WriteString("\t\"context\"\n")
	for _, imp := range extraImports {
		importBlock.WriteString("\t\"" + imp + "\"\n")
	}
	importBlock.WriteString("\n")
	importBlock.WriteString("\t\"" + modulePath + "/config\"\n")
	importBlock.WriteString("\t\"go.uber.org/fx\"\n")
	importBlock.WriteString("\t\"go.uber.org/zap\"\n")
	importBlock.WriteString(")\n")

	content := `package main

` + importBlock.String() + `
func main() {
	app := fx.New(
		fx.Provide(
			config.NewConfig,
			zap.NewProduction,
		),
		fx.Invoke(run),
	)

	app.Run()
}

func run(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting application")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping application")
			return nil
		},
	})
}
`
	return os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(content), 0644)
}

// getAllImportsFromFile parses a main.go and returns all import paths as strings
func getAllImportsFromFile(mainPath string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, mainPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var imports []string
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, importPath)
	}
	return imports, nil
}

// Feature: domain-module-management, Property 4: Import Preservation
// **Validates: Requirements 2.4**
func TestProperty_ImportPreservation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("adding domain import preserves all existing imports", prop.ForAll(
		func(modulePath string, extraImports []string) bool {
			tmpDir := t.TempDir()

			// Create main.go with base imports + extra imports
			if err := createMainGoWithImports(tmpDir, modulePath, extraImports); err != nil {
				t.Logf("failed to create main.go: %v", err)
				return false
			}

			// Record all imports before modification
			mainPath := filepath.Join(tmpDir, "cmd", "main.go")
			importsBefore, err := getAllImportsFromFile(mainPath)
			if err != nil {
				t.Logf("failed to get imports before: %v", err)
				return false
			}

			// Add domain module
			mgr := NewMainFileManager(tmpDir)
			if err := addUsecaseModuleForTest(mgr, modulePath); err != nil {
				t.Logf("failed to add domain module: %v", err)
				return false
			}

			// Record all imports after modification
			importsAfter, err := getAllImportsFromFile(mainPath)
			if err != nil {
				t.Logf("failed to get imports after: %v", err)
				return false
			}

			// Verify every original import is still present in the final imports
			afterSet := make(map[string]bool)
			for _, imp := range importsAfter {
				afterSet[imp] = true
			}

			for _, originalImport := range importsBefore {
				if !afterSet[originalImport] {
					t.Logf("original import %q not found after adding domain import", originalImport)
					return false
				}
			}

			return true
		},
		genValidModulePath(),
		genImportList(),
	))

	properties.TestingRun(t)
}

// Feature: domain-module-management, Property 4: Import Preservation
// **Validates: Requirements 2.4**
func TestProperty_ImportPreservationCount(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("final import count equals original count plus one (the domain import)", prop.ForAll(
		func(modulePath string, extraImports []string) bool {
			tmpDir := t.TempDir()

			// Create main.go with base imports + extra imports
			if err := createMainGoWithImports(tmpDir, modulePath, extraImports); err != nil {
				t.Logf("failed to create main.go: %v", err)
				return false
			}

			// Record import count before modification
			mainPath := filepath.Join(tmpDir, "cmd", "main.go")
			importsBefore, err := getAllImportsFromFile(mainPath)
			if err != nil {
				t.Logf("failed to get imports before: %v", err)
				return false
			}
			countBefore := len(importsBefore)

			// Add domain module
			mgr := NewMainFileManager(tmpDir)
			if err := addUsecaseModuleForTest(mgr, modulePath); err != nil {
				t.Logf("failed to add domain module: %v", err)
				return false
			}

			// Record import count after modification
			importsAfter, err := getAllImportsFromFile(mainPath)
			if err != nil {
				t.Logf("failed to get imports after: %v", err)
				return false
			}
			countAfter := len(importsAfter)

			// Verify the final import count equals original count + 1
			if countAfter != countBefore+1 {
				t.Logf("expected %d imports after, got %d (before: %d)", countBefore+1, countAfter, countBefore)
				return false
			}

			return true
		},
		genValidModulePath(),
		genImportList(),
	))

	properties.TestingRun(t)
}
