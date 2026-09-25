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

// Custom generators for usecase module property-based testing

// genUseCaseName generates valid use case names (PascalCase identifiers)
func genUseCaseName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{2,20}$`)
}

// genUseCaseNameWithMultipleWords generates use case names with multiple words
func genUseCaseNameWithMultipleWords() gopter.Gen {
	return gen.OneConstOf(
		"CreateUser",
		"GetUserById",
		"UpdateUserProfile",
		"DeleteAccount",
		"ProcessPayment",
		"SendNotification",
		"ValidateInput",
		"FetchData",
		"CalculateTotal",
		"GenerateReport",
	)
}

// genAnyUseCaseName generates any valid use case name
func genAnyUseCaseName() gopter.Gen {
	return gen.OneGenOf(
		genUseCaseName(),
		genUseCaseNameWithMultipleWords(),
	)
}

// createInitialModuleFile creates a usecase/module.go file with an initial provider
func createInitialModuleFile(projectDir, initialUseCaseName string) error {
	usecaseDir := filepath.Join(projectDir, "internal", "usecase")
	if err := os.MkdirAll(usecaseDir, 0755); err != nil {
		return err
	}

	modulePath := filepath.Join(usecaseDir, "module.go")
	content := `package usecase

import "go.uber.org/fx"

// Module provides use cases dependencies as an fx Module
var Module = fx.Module("usecase",
	fx.Provide(
		` + initialUseCaseName + `,
	),
)
`
	return os.WriteFile(modulePath, []byte(content), 0644)
}

// countProviderOccurrences counts how many times a provider appears in fx.Provide
func countProviderOccurrences(modulePath, providerName string) (int, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, modulePath, nil, parser.ParseComments)
	if err != nil {
		return 0, err
	}

	count := 0
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

		// Count occurrences of provider name
		for _, arg := range callExpr.Args {
			if argIdent, ok := arg.(*ast.Ident); ok && argIdent.Name == providerName {
				count++
			}
		}

		return true
	})

	return count, nil
}

// hasProviderInFxProvide checks if a provider exists in fx.Provide
func hasProviderInFxProvide(modulePath, providerName string) (bool, error) {
	count, err := countProviderOccurrences(modulePath, providerName)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Feature: usecase-module-management, Property 9: Provider Addition to Usecase Module
// **Validates: Requirements 4.2**
func TestProperty_ProviderAdditionToUsecaseModule(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("provider function New{Name}UseCase is added to fx.Provide", prop.ForAll(
		func(useCaseName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create initial module file with a different use case
			if err := createInitialModuleFile(tmpDir, "Initial"); err != nil {
				t.Logf("failed to create initial module file: %v", err)
				return false
			}

			// Add new provider
			manager := NewGenericModuleManager(tmpDir)
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			if err := manager.AddProvider(modulePath, "New"+useCaseName+"UseCase"); err != nil {
				t.Logf("failed to add provider: %v", err)
				return false
			}

			// Verify provider was added
			providerName := "New" + useCaseName + "UseCase"

			found, err := hasProviderInFxProvide(modulePath, providerName)
			if err != nil {
				t.Logf("failed to check for provider: %v", err)
				return false
			}

			if !found {
				t.Logf("provider %s not found in fx.Provide", providerName)
				return false
			}

			return true
		},
		genAnyUseCaseName(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 9: Provider Addition to Domain Module (Exactly Once)
// **Validates: Requirements 4.2**
func TestProperty_ProviderAppearsExactlyOnce(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("provider appears exactly once in fx.Provide", prop.ForAll(
		func(useCaseName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create initial module file
			if err := createInitialModuleFile(tmpDir, "Initial"); err != nil {
				t.Logf("failed to create initial module file: %v", err)
				return false
			}

			// Add provider
			manager := NewGenericModuleManager(tmpDir)
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")

			if err := manager.AddProvider(modulePath, "New"+useCaseName+"UseCase"); err != nil {
				t.Logf("failed to add provider: %v", err)
				return false
			}

			// Count occurrences
			providerName := "New" + useCaseName + "UseCase"

			count, err := countProviderOccurrences(modulePath, providerName)
			if err != nil {
				t.Logf("failed to count provider occurrences: %v", err)
				return false
			}

			if count != 1 {
				t.Logf("expected provider %s to appear exactly once, but found %d occurrences", providerName, count)
				return false
			}

			return true
		},
		genAnyUseCaseName(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 9: Provider Addition to Domain Module (Correct Format)
// **Validates: Requirements 4.2**
func TestProperty_ProviderHasCorrectFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("provider function name follows New{Name}UseCase format", prop.ForAll(
		func(useCaseName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create initial module file
			if err := createInitialModuleFile(tmpDir, "Initial"); err != nil {
				t.Logf("failed to create initial module file: %v", err)
				return false
			}

			// Add provider
			manager := NewGenericModuleManager(tmpDir)
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			if err := manager.AddProvider(modulePath, "New"+useCaseName+"UseCase"); err != nil {
				t.Logf("failed to add provider: %v", err)
				return false
			}

			// Read module file and verify format
			content, err := os.ReadFile(modulePath)
			if err != nil {
				t.Logf("failed to read module file: %v", err)
				return false
			}

			expectedProvider := "New" + useCaseName + "UseCase"
			if !strings.Contains(string(content), expectedProvider) {
				t.Logf("expected provider %s not found in module file", expectedProvider)
				return false
			}

			// Verify it starts with "New" and ends with "UseCase"
			if !strings.HasPrefix(expectedProvider, "New") {
				t.Logf("provider %s does not start with 'New'", expectedProvider)
				return false
			}

			if !strings.HasSuffix(expectedProvider, "UseCase") {
				t.Logf("provider %s does not end with 'UseCase'", expectedProvider)
				return false
			}

			return true
		},
		genAnyUseCaseName(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 9: Provider Addition to Domain Module (Multiple Additions)
// **Validates: Requirements 4.2**
func TestProperty_MultipleProviderAdditions(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("multiple different providers can be added sequentially", prop.ForAll(
		func(useCaseNames []string) bool {
			// Filter out duplicates and empty names
			uniqueNames := make(map[string]bool)
			var filteredNames []string
			for _, name := range useCaseNames {
				if name != "" && !uniqueNames[name] {
					uniqueNames[name] = true
					filteredNames = append(filteredNames, name)
				}
			}

			if len(filteredNames) == 0 {
				return true // Skip empty test case
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create initial module file with first use case
			if err := createInitialModuleFile(tmpDir, "New"+filteredNames[0]+"UseCase"); err != nil {
				t.Logf("failed to create initial module file: %v", err)
				return false
			}

			// Add remaining providers
			manager := NewGenericModuleManager(tmpDir)
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")

			for i := 1; i < len(filteredNames); i++ {
				if err := manager.AddProvider(modulePath, "New"+filteredNames[i]+"UseCase"); err != nil {
					t.Logf("failed to add provider %s: %v", filteredNames[i], err)
					return false
				}
			}

			// Verify all providers are present
			for _, name := range filteredNames {
				providerName := "New" + name + "UseCase"
				found, err := hasProviderInFxProvide(modulePath, providerName)
				if err != nil {
					t.Logf("failed to check for provider %s: %v", providerName, err)
					return false
				}
				if !found {
					t.Logf("provider %s not found after multiple additions", providerName)
					return false
				}
			}

			return true
		},
		gen.SliceOfN(5, genAnyUseCaseName()),
	))

	properties.TestingRun(t)
}

// genProviderList generates a list of 1-10 unique use case names for providers
func genProviderList() gopter.Gen {
	return gen.SliceOfN(10, genAnyUseCaseName()).
		Map(func(names []string) []string {
			// Remove duplicates
			uniqueNames := make(map[string]bool)
			var result []string
			for _, name := range names {
				if name != "" && !uniqueNames[name] {
					uniqueNames[name] = true
					result = append(result, name)
				}
			}
			// Ensure at least one provider
			if len(result) == 0 {
				result = append(result, "Initial")
			}
			return result
		})
}

// createModuleFileWithProviders creates a usecase/module.go file with multiple providers
func createModuleFileWithProviders(projectDir string, providerNames []string) error {
	usecaseDir := filepath.Join(projectDir, "internal", "usecase")
	if err := os.MkdirAll(usecaseDir, 0755); err != nil {
		return err
	}

	modulePath := filepath.Join(usecaseDir, "module.go")

	// Build provider list
	var providers strings.Builder
	for i, name := range providerNames {
		providers.WriteString("\t\tNew")
		providers.WriteString(name)
		providers.WriteString("UseCase,\n")
		if i < len(providerNames)-1 {
			// Add extra line for formatting (optional)
		}
	}

	content := `package usecase

import "go.uber.org/fx"

// Module provides use cases dependencies as an fx Module
var Module = fx.Module("usecase",
	fx.Provide(
` + providers.String() + `	),
)
`
	return os.WriteFile(modulePath, []byte(content), 0644)
}

// getAllProvidersFromModule extracts all provider names from fx.Provide
func getAllProvidersFromModule(modulePath string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, modulePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var providers []string
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

		// Extract all provider names
		for _, arg := range callExpr.Args {
			if argIdent, ok := arg.(*ast.Ident); ok {
				providers = append(providers, argIdent.Name)
			}
		}

		return true
	})

	return providers, nil
}

// Feature: usecase-module-management, Property 10: Provider Preservation
// **Validates: Requirements 4.3**
func TestProperty_ProviderPreservation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("adding new provider preserves all original providers", prop.ForAll(
		func(existingProviders []string, newUseCaseName string) bool {
			// Ensure new provider is different from existing ones
			for _, existing := range existingProviders {
				if existing == newUseCaseName {
					return true // Skip if duplicate
				}
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create module file with existing providers
			if err := createModuleFileWithProviders(tmpDir, existingProviders); err != nil {
				t.Logf("failed to create module file with providers: %v", err)
				return false
			}

			// Get original provider list
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			originalProviders, err := getAllProvidersFromModule(modulePath)
			if err != nil {
				t.Logf("failed to get original providers: %v", err)
				return false
			}

			// Add new provider
			manager := NewGenericModuleManager(tmpDir)
			if err := manager.AddProvider(modulePath, "New"+newUseCaseName+"UseCase"); err != nil {
				t.Logf("failed to add new provider: %v", err)
				return false
			}

			// Get final provider list
			finalProviders, err := getAllProvidersFromModule(modulePath)
			if err != nil {
				t.Logf("failed to get final providers: %v", err)
				return false
			}

			// Verify all original providers are still present
			originalSet := make(map[string]bool)
			for _, provider := range originalProviders {
				originalSet[provider] = true
			}

			finalSet := make(map[string]bool)
			for _, provider := range finalProviders {
				finalSet[provider] = true
			}

			// Check that original providers are a subset of final providers
			for originalProvider := range originalSet {
				if !finalSet[originalProvider] {
					t.Logf("original provider %s not found in final provider list", originalProvider)
					return false
				}
			}

			// Verify the new provider was added
			newProviderName := "New" + newUseCaseName + "UseCase"
			if !finalSet[newProviderName] {
				t.Logf("new provider %s not found in final provider list", newProviderName)
				return false
			}

			return true
		},
		genProviderList(),
		genAnyUseCaseName(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 10: Provider Preservation (Count Verification)
// **Validates: Requirements 4.3**
func TestProperty_ProviderPreservationCount(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("final provider count equals original count plus one", prop.ForAll(
		func(existingProviders []string, newUseCaseName string) bool {
			// Ensure new provider is different from existing ones
			for _, existing := range existingProviders {
				if existing == newUseCaseName {
					return true // Skip if duplicate
				}
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create module file with existing providers
			if err := createModuleFileWithProviders(tmpDir, existingProviders); err != nil {
				t.Logf("failed to create module file with providers: %v", err)
				return false
			}

			// Get original provider count
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			originalProviders, err := getAllProvidersFromModule(modulePath)
			if err != nil {
				t.Logf("failed to get original providers: %v", err)
				return false
			}
			originalCount := len(originalProviders)

			// Add new provider
			manager := NewGenericModuleManager(tmpDir)
			if err := manager.AddProvider(modulePath, newUseCaseName); err != nil {
				t.Logf("failed to add new provider: %v", err)
				return false
			}

			// Get final provider count
			finalProviders, err := getAllProvidersFromModule(modulePath)
			if err != nil {
				t.Logf("failed to get final providers: %v", err)
				return false
			}
			finalCount := len(finalProviders)

			// Verify count increased by exactly one
			if finalCount != originalCount+1 {
				t.Logf("expected final count to be %d, but got %d", originalCount+1, finalCount)
				return false
			}

			return true
		},
		genProviderList(),
		genAnyUseCaseName(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 10: Provider Preservation (Order Independence)
// **Validates: Requirements 4.3**
func TestProperty_ProviderPreservationOrderIndependent(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("provider preservation works regardless of existing provider order", prop.ForAll(
		func(existingProviders []string, newUseCaseName string) bool {
			// Ensure new provider is different from existing ones
			for _, existing := range existingProviders {
				if existing == newUseCaseName {
					return true // Skip if duplicate
				}
			}

			if len(existingProviders) < 2 {
				return true // Need at least 2 providers to test order independence
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create module file with existing providers
			if err := createModuleFileWithProviders(tmpDir, existingProviders); err != nil {
				t.Logf("failed to create module file with providers: %v", err)
				return false
			}

			// Add new provider
			manager := NewGenericModuleManager(tmpDir)
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			if err := manager.AddProvider(modulePath, newUseCaseName); err != nil {
				t.Logf("failed to add new provider: %v", err)
				return false
			}

			// Get final provider list
			finalProviders, err := getAllProvidersFromModule(modulePath)
			if err != nil {
				t.Logf("failed to get final providers: %v", err)
				return false
			}

			// Verify all original providers are present (order doesn't matter)
			finalSet := make(map[string]bool)
			for _, provider := range finalProviders {
				finalSet[provider] = true
			}

			for _, original := range existingProviders {
				expectedName := "New" + original + "UseCase"
				if !finalSet[expectedName] {
					t.Logf("original provider %s not found in final list", expectedName)
					return false
				}
			}

			return true
		},
		genProviderList(),
		genAnyUseCaseName(),
	))

	properties.TestingRun(t)
}

// getModuleStructure extracts the fx.Module structure from the module file
func getModuleStructure(modulePath string) (moduleName string, hasFxModule bool, err error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, modulePath, nil, parser.ParseComments)
	if err != nil {
		return "", false, err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for fx.Module call
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if this is fx.Module
		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok || ident.Name != "fx" || selExpr.Sel.Name != "Module" {
			return true
		}

		hasFxModule = true

		// Extract module name (first argument)
		if len(callExpr.Args) > 0 {
			if basicLit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
				// Remove quotes from string literal
				moduleName = strings.Trim(basicLit.Value, `"`)
			}
		}

		return false
	})

	return moduleName, hasFxModule, nil
}

// Feature: usecase-module-management, Property 11: Module Structure Preservation
// **Validates: Requirements 4.6**
func TestProperty_ModuleStructurePreservation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("fx.Module wrapper and usecase name unchanged after adding provider", prop.ForAll(
		func(existingProviders []string, newUseCaseName string) bool {
			// Ensure new provider is different from existing ones
			for _, existing := range existingProviders {
				if existing == newUseCaseName {
					return true // Skip if duplicate
				}
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create module file with existing providers
			if err := createModuleFileWithProviders(tmpDir, existingProviders); err != nil {
				t.Logf("failed to create module file with providers: %v", err)
				return false
			}

			// Get original module structure
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			originalModuleName, originalHasFxModule, err := getModuleStructure(modulePath)
			if err != nil {
				t.Logf("failed to get original module structure: %v", err)
				return false
			}

			// Add new provider
			manager := NewGenericModuleManager(tmpDir)
			if err := manager.AddProvider(modulePath, newUseCaseName); err != nil {
				t.Logf("failed to add new provider: %v", err)
				return false
			}

			// Get final module structure
			finalModuleName, finalHasFxModule, err := getModuleStructure(modulePath)
			if err != nil {
				t.Logf("failed to get final module structure: %v", err)
				return false
			}

			// Verify fx.Module wrapper is still present
			if !originalHasFxModule || !finalHasFxModule {
				t.Logf("fx.Module wrapper missing (original: %v, final: %v)", originalHasFxModule, finalHasFxModule)
				return false
			}

			// Verify module name is unchanged
			if originalModuleName != finalModuleName {
				t.Logf("module name changed from %q to %q", originalModuleName, finalModuleName)
				return false
			}

			// Verify module name is "usecase"
			if finalModuleName != "usecase" {
				t.Logf("expected module name to be 'usecase', but got %q", finalModuleName)
				return false
			}

			return true
		},
		genProviderList(),
		genAnyUseCaseName(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 11: Module Structure Preservation (AST Comparison)
// **Validates: Requirements 4.6**
func TestProperty_ModuleStructurePreservationAST(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("fx.Module AST structure preserved after adding provider", prop.ForAll(
		func(existingProviders []string, newUseCaseName string) bool {
			// Ensure new provider is different from existing ones
			for _, existing := range existingProviders {
				if existing == newUseCaseName {
					return true // Skip if duplicate
				}
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create module file with existing providers
			if err := createModuleFileWithProviders(tmpDir, existingProviders); err != nil {
				t.Logf("failed to create module file with providers: %v", err)
				return false
			}

			// Parse original AST
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			fset := token.NewFileSet()
			originalNode, err := parser.ParseFile(fset, modulePath, nil, parser.ParseComments)
			if err != nil {
				t.Logf("failed to parse original module file: %v", err)
				return false
			}

			// Count original fx.Module calls
			originalModuleCount := 0
			ast.Inspect(originalNode, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := selExpr.X.(*ast.Ident)
				if ok && ident.Name == "fx" && selExpr.Sel.Name == "Module" {
					originalModuleCount++
				}

				return true
			})

			// Add new provider
			manager := NewGenericModuleManager(tmpDir)
			if err := manager.AddProvider(modulePath, newUseCaseName); err != nil {
				t.Logf("failed to add new provider: %v", err)
				return false
			}

			// Parse final AST
			fset = token.NewFileSet()
			finalNode, err := parser.ParseFile(fset, modulePath, nil, parser.ParseComments)
			if err != nil {
				t.Logf("failed to parse final module file: %v", err)
				return false
			}

			// Count final fx.Module calls
			finalModuleCount := 0
			ast.Inspect(finalNode, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := selExpr.X.(*ast.Ident)
				if ok && ident.Name == "fx" && selExpr.Sel.Name == "Module" {
					finalModuleCount++
				}

				return true
			})

			// Verify fx.Module count is unchanged (should be exactly 1)
			if originalModuleCount != finalModuleCount {
				t.Logf("fx.Module count changed from %d to %d", originalModuleCount, finalModuleCount)
				return false
			}

			if finalModuleCount != 1 {
				t.Logf("expected exactly 1 fx.Module call, but found %d", finalModuleCount)
				return false
			}

			return true
		},
		genProviderList(),
		genAnyUseCaseName(),
	))

	properties.TestingRun(t)
}

// Feature: usecase-module-management, Property 11: Module Structure Preservation (Variable Declaration)
// **Validates: Requirements 4.6**
func TestProperty_ModuleStructurePreservationVariableDeclaration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Module variable declaration preserved after adding provider", prop.ForAll(
		func(existingProviders []string, newUseCaseName string) bool {
			// Ensure new provider is different from existing ones
			for _, existing := range existingProviders {
				if existing == newUseCaseName {
					return true // Skip if duplicate
				}
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create module file with existing providers
			if err := createModuleFileWithProviders(tmpDir, existingProviders); err != nil {
				t.Logf("failed to create module file with providers: %v", err)
				return false
			}

			// Check original file has "var Module"
			modulePath := filepath.Join(tmpDir, "internal", "usecase", "module.go")
			originalContent, err := os.ReadFile(modulePath)
			if err != nil {
				t.Logf("failed to read original module file: %v", err)
				return false
			}

			if !strings.Contains(string(originalContent), "var Module") {
				t.Logf("original file does not contain 'var Module' declaration")
				return false
			}

			// Add new provider
			manager := NewGenericModuleManager(tmpDir)
			if err := manager.AddProvider(modulePath, newUseCaseName); err != nil {
				t.Logf("failed to add new provider: %v", err)
				return false
			}

			// Check final file still has "var Module"
			finalContent, err := os.ReadFile(modulePath)
			if err != nil {
				t.Logf("failed to read final module file: %v", err)
				return false
			}

			if !strings.Contains(string(finalContent), "var Module") {
				t.Logf("final file does not contain 'var Module' declaration")
				return false
			}

			// Verify the module name "usecase" is still in the fx.Module call
			if !strings.Contains(string(finalContent), `fx.Module("usecase"`) {
				t.Logf("final file does not contain fx.Module(\"usecase\")")
				return false
			}

			return true
		},
		genProviderList(),
		genAnyUseCaseName(),
	))

	properties.TestingRun(t)
}
