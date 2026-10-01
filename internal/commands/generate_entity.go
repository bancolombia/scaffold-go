package commands

import (
	"fmt"
	"strings"

	"github.com/bancolombia/scaffold-go/internal/filesystem"
	"github.com/bancolombia/scaffold-go/internal/generator"
	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewGenerateEntityCommand() *cobra.Command {
	var fields []string

	cmd := &cobra.Command{
		Use:   "entity [name]",
		Short: "Generate a domain entity",
		Long: `Generate a domain entity with fields and a repository interface.

This command creates two files in the domain layer:
  1. Entity file with struct definition and Validate method
  2. Repository interface with CRUD methods (Create, GetByID, Update, Delete, List)

The entity is pure business logic with no external dependencies.
Field types support Go primitives and standard library types.

Field format: Name:type
  • Supported types: string, int, int64, float64, bool, time.Time, etc.
  • Example: Name:string,Email:string,Age:int,CreatedAt:time.Time`,
		Args: cobra.ExactArgs(1),
		Example: `  # Generate a User entity with basic fields
  scaffold-go generate entity User --fields Name:string,Email:string,Age:int

  # Generate a Product entity with pricing
  scaffold-go gen entity Product --fields Name:string,Price:float64,Stock:int

  # Generate an Order entity with timestamp
  scaffold-go generate entity Order --fields UserID:string,Total:float64,CreatedAt:time.Time

  # Use short flag notation
  scaffold-go gen entity Customer -f Name:string,Email:string,Phone:string`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			projectDir := "."

			logger := GetLogger()
			logger.Info("Starting entity generation",
				zap.String("entity_name", name),
				zap.Strings("fields", fields),
				zap.String("project_dir", projectDir),
			)

			validator := GetValidator()
			logger.Debug("Validating project initialization")
			if err := validator.ValidateProjectInitialized(projectDir); err != nil {
				logger.Error("Project not initialized",
					zap.Error(err),
					zap.String("project_dir", projectDir),
				)
				return WrapValidationError("project", projectDir, err)
			}

			logger.Debug("Initializing template engine")
			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				logger.Error("Failed to initialize template engine",
					zap.Error(err),
				)
				return WrapTemplateError("entity", "initialization", err)
			}

			gen := generator.NewEntityGenerator(name, fields, projectDir, tmpl)

			logger.Debug("Validating entity parameters")
			if err := gen.Validate(); err != nil {
				logger.Error("Entity validation failed",
					zap.Error(err),
					zap.String("entity_name", name),
				)
				return WrapValidationError("entity", name, err)
			}

			logger.Debug("Generating entity files")
			files, err := gen.Generate()
			if err != nil {
				logger.Error("Entity generation failed",
					zap.Error(err),
					zap.String("entity_name", name),
				)
				return WrapGenerationError("entity", "generation", err)
			}
			logger.Debug("Entity files generated",
				zap.Int("file_count", len(files)),
			)

			fs := filesystem.New()
			operations := make([]filesystem.FileOperation, len(files))
			for i, f := range files {
				content := []byte(f.Content)
				if strings.HasSuffix(f.Path, ".go") {
					logger.Debug("Formatting Go code", zap.String("file", f.Path))
					content, err = fs.FormatGoCode(content)
					if err != nil {
						logger.Error("Failed to format Go code",
							zap.Error(err),
							zap.String("file", f.Path),
						)
						return WrapFileSystemError("format", f.Path, err)
					}
				}
				operations[i] = filesystem.FileOperation{
					Path:    f.Path,
					Content: content,
				}
			}

			logger.Debug("Writing files to disk")
			if err := fs.WriteFiles(operations); err != nil {
				logger.Error("Failed to write files",
					zap.Error(err),
					zap.String("project_dir", projectDir),
				)
				return WrapFileSystemError("write", projectDir, err)
			}

			logger.Info("Entity generated successfully",
				zap.String("entity_name", name),
				zap.Int("files_created", len(files)),
				zap.Strings("file_paths", func() []string {
					paths := make([]string, len(files))
					for i, f := range files {
						paths[i] = f.Path
					}
					return paths
				}()),
			)
			fmt.Printf("✓ Entity '%s' generated\n", name)
			for _, f := range files {
				fmt.Printf("  - %s\n", f.Path)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&fields, "fields", "f", []string{}, "Entity fields in format Name:type (comma-separated). Example: Name:string,Email:string,Age:int")
	_ = cmd.MarkFlagRequired("fields")

	return cmd
}
