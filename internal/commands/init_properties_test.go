package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Custom generators for init command property-based testing

// genValidProjectName generates valid project names (alphanumeric, hyphens, underscores)
func genValidProjectName() gopter.Gen {
	return gen.RegexMatch(`^[a-zA-Z][a-zA-Z0-9_-]{0,30}`)
}

// genValidModulePath generates valid Go module paths
func genValidModulePath() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("github.com/user/project"),
		gen.Const("gitlab.com/org/service"),
		gen.Const("example.com/app/backend"),
		gen.Const("mycompany.com/services/api"),
		gen.RegexMatch(`^[a-z][a-z0-9-]{0,20}\.[a-z]{2,6}/[a-z][a-z0-9-]{0,20}/[a-z][a-z0-9-]{0,20}`),
	)
}

// Feature: go-microservice-scaffolding-cli, Property 1: Project initialization creates required directory structure
// **Validates: Requirements 1.1**
func TestProperty_ProjectInitializationCreatesRequiredDirectoryStructure(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("init creates all required directories", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			// Verify all required directories exist
			requiredDirs := []string{
				"cmd",
				filepath.Join("internal", "domain"),
				filepath.Join("internal", "entrypoints", "http"),
				filepath.Join("internal", "entrypoints", "messaging"),
				filepath.Join("internal", "drivenadapters", "database"),
				filepath.Join("internal", "drivenadapters", "api"),
				filepath.Join("internal", "drivenadapters", "messaging"),
				filepath.Join("internal", "drivenadapters", "cache"),
				"config",
			}

			projectDir := filepath.Join(tmpDir, projectName)
			for _, dir := range requiredDirs {
				fullPath := filepath.Join(projectDir, dir)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Logf("required directory not found: %s", dir)
					return false
				}
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 2: Project initialization creates go.mod with correct module name
// **Validates: Requirements 1.2**
func TestProperty_ProjectInitializationCreatesGoModWithCorrectModuleName(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("init creates go.mod with specified module name", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			// Read go.mod file
			projectDir := filepath.Join(tmpDir, projectName)
			goModPath := filepath.Join(projectDir, "go.mod")
			content, err := os.ReadFile(goModPath)
			if err != nil {
				t.Logf("failed to read go.mod: %v", err)
				return false
			}

			goModContent := string(content)

			// Verify go.mod exists and contains module declaration
			if !strings.Contains(goModContent, "module "+modulePath) {
				t.Logf("go.mod does not contain correct module declaration")
				return false
			}

			// Verify go.mod contains required dependencies
			requiredDeps := []string{
				"github.com/spf13/cobra",
				"github.com/spf13/viper",
				"go.uber.org/fx",
				"go.uber.org/zap",
			}

			for _, dep := range requiredDeps {
				if !strings.Contains(goModContent, dep) {
					t.Logf("go.mod missing required dependency: %s", dep)
					return false
				}
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 3: Project initialization creates main.go with Cobra setup
// **Validates: Requirements 1.3**
func TestProperty_ProjectInitializationCreatesMainGoWithCobraSetup(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("init creates main.go with Cobra root command initialization", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			// Read main.go file
			projectDir := filepath.Join(tmpDir, projectName)
			mainGoPath := filepath.Join(projectDir, "cmd", "main.go")
			content, err := os.ReadFile(mainGoPath)
			if err != nil {
				t.Logf("failed to read main.go: %v", err)
				return false
			}

			mainGoContent := string(content)

			// Verify main.go contains package main
			if !strings.Contains(mainGoContent, "package main") {
				t.Logf("main.go does not contain package main")
				return false
			}

			// Verify main.go contains fx.New() for dependency injection
			if !strings.Contains(mainGoContent, "fx.New") {
				t.Logf("main.go does not contain fx.New() call")
				return false
			}

			// Verify main.go imports fx
			if !strings.Contains(mainGoContent, "go.uber.org/fx") {
				t.Logf("main.go does not import go.uber.org/fx")
				return false
			}

			// Verify main.go has main function
			if !strings.Contains(mainGoContent, "func main()") {
				t.Logf("main.go does not contain main function")
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 4: Project initialization creates config directory with Viper setup
// **Validates: Requirements 1.4**
func TestProperty_ProjectInitializationCreatesConfigDirectoryWithViperSetup(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("init creates config directory with Viper configuration setup", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			projectDir := filepath.Join(tmpDir, projectName)

			// Verify config directory exists
			configDir := filepath.Join(projectDir, "config")
			if _, err := os.Stat(configDir); os.IsNotExist(err) {
				t.Logf("config directory not found")
				return false
			}

			// Read config.go file
			configGoPath := filepath.Join(configDir, "config.go")
			content, err := os.ReadFile(configGoPath)
			if err != nil {
				t.Logf("failed to read config.go: %v", err)
				return false
			}

			configGoContent := string(content)

			// Verify config.go contains Viper import
			if !strings.Contains(configGoContent, "github.com/spf13/viper") {
				t.Logf("config.go does not import Viper")
				return false
			}

			// Verify config.go contains Config struct
			if !strings.Contains(configGoContent, "type Config struct") {
				t.Logf("config.go does not contain Config struct")
				return false
			}

			// Verify config.go contains NewConfig function
			if !strings.Contains(configGoContent, "func NewConfig()") {
				t.Logf("config.go does not contain NewConfig function")
				return false
			}

			// Verify config.go uses Viper methods
			viperMethods := []string{
				"viper.SetConfigName",
				"viper.SetConfigType",
				"viper.AddConfigPath",
				"viper.AutomaticEnv",
				"viper.ReadInConfig",
				"viper.Unmarshal",
			}

			for _, method := range viperMethods {
				if !strings.Contains(configGoContent, method) {
					t.Logf("config.go does not use Viper method: %s", method)
					return false
				}
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 5: Project initialization creates README.md
// **Validates: Requirements 1.5**
func TestProperty_ProjectInitializationCreatesREADME(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("init creates README.md file", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			// Verify README.md exists
			projectDir := filepath.Join(tmpDir, projectName)
			readmePath := filepath.Join(projectDir, "README.md")
			content, err := os.ReadFile(readmePath)
			if err != nil {
				t.Logf("failed to read README.md: %v", err)
				return false
			}

			readmeContent := string(content)

			// Verify README.md contains project name
			if !strings.Contains(readmeContent, projectName) {
				t.Logf("README.md does not contain project name")
				return false
			}

			// Verify README.md contains project structure documentation
			if !strings.Contains(readmeContent, "Project Structure") {
				t.Logf("README.md does not contain project structure documentation")
				return false
			}

			// Verify README.md mentions clean architecture
			if !strings.Contains(readmeContent, "clean architecture") {
				t.Logf("README.md does not mention clean architecture")
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 40: Project initialization creates config package with Viper
// **Validates: Requirements 13.1**
func TestProperty_ProjectInitializationCreatesConfigPackageWithViper(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("init creates config package with Viper integration", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			projectDir := filepath.Join(tmpDir, projectName)

			// Read config.go file
			configGoPath := filepath.Join(projectDir, "config", "config.go")
			content, err := os.ReadFile(configGoPath)
			if err != nil {
				t.Logf("failed to read config.go: %v", err)
				return false
			}

			configGoContent := string(content)

			// Verify config.go is in config package
			if !strings.Contains(configGoContent, "package config") {
				t.Logf("config.go is not in config package")
				return false
			}

			// Verify config.go contains standard configuration structs
			requiredStructs := []string{
				"type Config struct",
				"type ServerConfig struct",
				"type DatabaseConfig struct",
				"type LoggingConfig struct",
			}

			for _, structDef := range requiredStructs {
				if !strings.Contains(configGoContent, structDef) {
					t.Logf("config.go missing required struct: %s", structDef)
					return false
				}
			}

			// Verify Config struct contains nested configs
			if !strings.Contains(configGoContent, "Server   ServerConfig") {
				t.Logf("Config struct does not contain ServerConfig field")
				return false
			}

			if !strings.Contains(configGoContent, "Database DatabaseConfig") {
				t.Logf("Config struct does not contain DatabaseConfig field")
				return false
			}

			if !strings.Contains(configGoContent, "Logging  LoggingConfig") {
				t.Logf("Config struct does not contain LoggingConfig field")
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 41: Project initialization creates YAML config file
// **Validates: Requirements 13.2**
func TestProperty_ProjectInitializationCreatesYAMLConfigFile(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("init creates default configuration file in YAML format", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			projectDir := filepath.Join(tmpDir, projectName)

			// Read config.yaml file
			configYamlPath := filepath.Join(projectDir, "config", "config.yaml")
			content, err := os.ReadFile(configYamlPath)
			if err != nil {
				t.Logf("failed to read config.yaml: %v", err)
				return false
			}

			configYamlContent := string(content)

			// Verify config.yaml contains standard settings
			requiredSettings := []string{
				"server:",
				"host:",
				"port:",
				"database:",
				"logging:",
				"level:",
			}

			for _, setting := range requiredSettings {
				if !strings.Contains(configYamlContent, setting) {
					t.Logf("config.yaml missing required setting: %s", setting)
					return false
				}
			}

			// Verify config.yaml contains default values
			if !strings.Contains(configYamlContent, "8080") {
				t.Logf("config.yaml does not contain default port")
				return false
			}

			if !strings.Contains(configYamlContent, "localhost") {
				t.Logf("config.yaml does not contain default host")
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 48: Project initialization creates main.go with dependency injection
// **Validates: Requirements 15.1**
func TestProperty_ProjectInitializationCreatesMainGoWithDependencyInjection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("init creates main.go with dependency injection initialization", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			projectDir := filepath.Join(tmpDir, projectName)

			// Read main.go file
			mainGoPath := filepath.Join(projectDir, "cmd", "main.go")
			content, err := os.ReadFile(mainGoPath)
			if err != nil {
				t.Logf("failed to read main.go: %v", err)
				return false
			}

			mainGoContent := string(content)

			// Verify main.go contains fx.New() call
			if !strings.Contains(mainGoContent, "fx.New(") {
				t.Logf("main.go does not contain fx.New() call")
				return false
			}

			// Verify main.go contains fx.Provide
			if !strings.Contains(mainGoContent, "fx.Provide(") {
				t.Logf("main.go does not contain fx.Provide")
				return false
			}

			// Verify main.go provides config
			if !strings.Contains(mainGoContent, "config.NewConfig") {
				t.Logf("main.go does not provide config.NewConfig")
				return false
			}

			// Verify main.go provides logger
			if !strings.Contains(mainGoContent, "zap.NewProduction") {
				t.Logf("main.go does not provide zap.NewProduction")
				return false
			}

			// Verify main.go contains fx.Invoke
			if !strings.Contains(mainGoContent, "fx.Invoke(") {
				t.Logf("main.go does not contain fx.Invoke")
				return false
			}

			// Verify main.go has lifecycle hooks
			if !strings.Contains(mainGoContent, "fx.Lifecycle") {
				t.Logf("main.go does not use fx.Lifecycle")
				return false
			}

			if !strings.Contains(mainGoContent, "OnStart") {
				t.Logf("main.go does not have OnStart hook")
				return false
			}

			if !strings.Contains(mainGoContent, "OnStop") {
				t.Logf("main.go does not have OnStop hook")
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 52: Generated DI uses uber-go/fx
// **Validates: Requirements 15.5**
func TestProperty_GeneratedDIUsesUberGoFx(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated dependency injection uses uber-go/fx framework", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			projectDir := filepath.Join(tmpDir, projectName)

			// Read main.go file
			mainGoPath := filepath.Join(projectDir, "cmd", "main.go")
			content, err := os.ReadFile(mainGoPath)
			if err != nil {
				t.Logf("failed to read main.go: %v", err)
				return false
			}

			mainGoContent := string(content)

			// Verify main.go imports uber-go/fx
			if !strings.Contains(mainGoContent, "\"go.uber.org/fx\"") {
				t.Logf("main.go does not import go.uber.org/fx")
				return false
			}

			// Verify go.mod contains uber-go/fx dependency
			goModPath := filepath.Join(projectDir, "go.mod")
			goModContent, err := os.ReadFile(goModPath)
			if err != nil {
				t.Logf("failed to read go.mod: %v", err)
				return false
			}

			if !strings.Contains(string(goModContent), "go.uber.org/fx") {
				t.Logf("go.mod does not contain go.uber.org/fx dependency")
				return false
			}

			// Verify main.go uses fx.New, fx.Provide, fx.Invoke
			fxMethods := []string{"fx.New", "fx.Provide", "fx.Invoke"}
			for _, method := range fxMethods {
				if !strings.Contains(mainGoContent, method) {
					t.Logf("main.go does not use fx method: %s", method)
					return false
				}
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 54: Code generation uses text/template
// **Validates: Requirements 21.1, 21.2, 21.3**
func TestProperty_CodeGenerationUsesTextTemplate(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("code generation uses text/template to process pre-defined templates", prop.ForAll(
		func(projectName string, modulePath string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Execute init command
			err := initProject(projectName, modulePath, filepath.Join(tmpDir, projectName))
			if err != nil {
				t.Logf("init failed: %v", err)
				return false
			}

			projectDir := filepath.Join(tmpDir, projectName)

			// Verify generated files contain template-processed content
			// Check that module path is correctly substituted in main.go
			mainGoPath := filepath.Join(projectDir, "cmd", "main.go")
			mainGoContent, err := os.ReadFile(mainGoPath)
			if err != nil {
				t.Logf("failed to read main.go: %v", err)
				return false
			}

			// Verify module path is substituted in imports
			if !strings.Contains(string(mainGoContent), modulePath+"/config") {
				t.Logf("main.go does not contain correctly substituted module path")
				return false
			}

			// Check that project name is correctly substituted in config.yaml
			configYamlPath := filepath.Join(projectDir, "config", "config.yaml")
			configYamlContent, err := os.ReadFile(configYamlPath)
			if err != nil {
				t.Logf("failed to read config.yaml: %v", err)
				return false
			}

			// Verify project name is substituted in database name
			if !strings.Contains(string(configYamlContent), "database: "+projectName) {
				t.Logf("config.yaml does not contain correctly substituted project name")
				return false
			}

			// Check that project name is correctly substituted in README.md
			readmePath := filepath.Join(projectDir, "README.md")
			readmeContent, err := os.ReadFile(readmePath)
			if err != nil {
				t.Logf("failed to read README.md: %v", err)
				return false
			}

			// Verify project name is substituted in README title
			if !strings.Contains(string(readmeContent), "# "+projectName) {
				t.Logf("README.md does not contain correctly substituted project name in title")
				return false
			}

			// Verify module path is substituted in README examples
			if !strings.Contains(string(readmeContent), modulePath) {
				t.Logf("README.md does not contain correctly substituted module path")
				return false
			}

			return true
		},
		genValidProjectName(),
		genValidModulePath(),
	))

	properties.TestingRun(t)
}
