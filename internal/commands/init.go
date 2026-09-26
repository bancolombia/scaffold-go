package commands

import (
	"fmt"
	"path/filepath"

	"github.com/gabheadz/scaffold-go/internal/filesystem"
	"github.com/gabheadz/scaffold-go/internal/template"
	"github.com/gabheadz/scaffold-go/internal/validator"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewInitCommand() *cobra.Command {
	var (
		modulePath string
		outputDir  string
	)

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Go microservice project",
		Long: `Initialize a new Go microservice project with clean architecture structure.

This command creates a complete project structure with:
  • Directory layout for domain, entrypoints, and drivenadapters layers
  • go.mod file with the specified module path
  • main.go with Cobra CLI and uber-go/fx dependency injection setup
  • Configuration package with Viper integration
  • Default YAML configuration file
  • README.md with project documentation

The generated project follows clean architecture principles where:
  • Domain layer contains pure business logic (entities, use cases, interfaces)
  • EntryPoints layer handles external requests (HTTP, messaging)
  • DrivenAdapters layer implements infrastructure (databases, APIs, caching)`,
		Args: cobra.ExactArgs(1),
		Example: `  # Initialize with default module path (same as project name)
  scaffold-go init my-service

  # Initialize with custom module path
  scaffold-go init my-service --module github.com/user/my-service

  # Initialize in a specific directory
  scaffold-go init my-service --module github.com/user/my-service --output ./projects

  # Initialize with short flags
  scaffold-go init my-service -m github.com/user/my-service -o ./projects`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			v := validator.New()
			if err := v.ValidateProjectName(projectName); err != nil {
				return WrapValidationError("project-name", projectName, err)
			}

			if outputDir == "" {
				outputDir = "."
			}
			projectDir := filepath.Join(outputDir, projectName)

			if err := v.ValidateDirectoryNotExists(projectDir); err != nil {
				return WrapValidationError("project-directory", projectDir, err)
			}

			if modulePath == "" {
				modulePath = projectName
			}

			return initProject(projectName, modulePath, projectDir)
		},
	}

	cmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path (e.g., github.com/user/project). Defaults to project name if not specified")
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory where the project will be created. Defaults to current directory")

	return cmd
}

func initProject(name, modulePath, projectDir string) error {
	logger := GetLogger()
	logger.Info("Starting project initialization",
		zap.String("project_name", name),
		zap.String("module_path", modulePath),
		zap.String("output_dir", projectDir),
	)

	fs := filesystem.New()
	tmpl, err := template.NewTemplateEngine()
	if err != nil {
		logger.Error("Failed to initialize template engine",
			zap.Error(err),
			zap.String("project_name", name),
		)
		return WrapTemplateError("project", "initialization", err)
	}
	logger.Debug("Template engine initialized successfully")

	dirs := []string{
		filepath.Join(projectDir, "cmd"),
		filepath.Join(projectDir, "internal", "domain"),
		filepath.Join(projectDir, "internal", "entrypoints", "http"),
		filepath.Join(projectDir, "internal", "entrypoints", "messaging"),
		filepath.Join(projectDir, "internal", "drivenadapters", "database"),
		filepath.Join(projectDir, "internal", "drivenadapters", "api"),
		filepath.Join(projectDir, "internal", "drivenadapters", "messaging"),
		filepath.Join(projectDir, "internal", "drivenadapters", "cache"),
		filepath.Join(projectDir, "internal", "infrastructure", "messaging"),
		filepath.Join(projectDir, "config"),
	}

	logger.Debug("Creating directory structure",
		zap.Int("directory_count", len(dirs)),
	)
	for _, dir := range dirs {
		logger.Debug("Creating directory", zap.String("path", dir))
		if err := fs.CreateDirectory(dir); err != nil {
			logger.Error("Failed to create directory",
				zap.Error(err),
				zap.String("path", dir),
			)
			return WrapFileSystemError("create", dir, err)
		}
	}
	logger.Debug("Directory structure created successfully")

	data := map[string]interface{}{
		"Name":       name,
		"ModulePath": modulePath,
	}

	files := map[string]string{
		"project/go.mod.tmpl":      filepath.Join(projectDir, "go.mod"),
		"project/main.go.tmpl":     filepath.Join(projectDir, "cmd", "main.go"),
		"project/config.go.tmpl":   filepath.Join(projectDir, "config", "config.go"),
		"project/config.yaml.tmpl": filepath.Join(projectDir, "config", "config.yaml"),
		"project/README.md.tmpl":   filepath.Join(projectDir, "README.md"),
		"project/Makefile.tmpl":    filepath.Join(projectDir, "Makefile"),
	}

	logger.Debug("Generating project files",
		zap.Int("file_count", len(files)),
	)
	operations := []filesystem.FileOperation{}
	for tmplName, path := range files {
		logger.Debug("Processing template",
			zap.String("template", tmplName),
			zap.String("output", path),
		)
		content, err := tmpl.Execute(tmplName, data)
		if err != nil {
			logger.Error("Failed to execute template",
				zap.Error(err),
				zap.String("template", tmplName),
			)
			return WrapTemplateError(tmplName, "execution", err)
		}
		operations = append(operations, filesystem.FileOperation{
			Path:    path,
			Content: []byte(content),
		})
	}

	logger.Debug("Writing files to disk")
	if err := fs.WriteFiles(operations); err != nil {
		logger.Error("Failed to write files",
			zap.Error(err),
			zap.String("project_dir", projectDir),
		)
		return WrapFileSystemError("write", projectDir, err)
	}

	logger.Info("Project initialized successfully",
		zap.String("project_name", name),
		zap.String("project_path", projectDir),
		zap.String("module_path", modulePath),
		zap.Int("files_created", len(files)),
		zap.Int("directories_created", len(dirs)),
	)
	fmt.Printf("✓ Project '%s' created at %s\n", name, projectDir)
	fmt.Printf("✓ Module path: %s\n", modulePath)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectDir)
	fmt.Println("  go mod tidy")
	fmt.Println("  scaffold-go generate entity User --fields Name:string,Email:string")

	return nil
}
