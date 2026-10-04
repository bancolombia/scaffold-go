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

func NewGenerateInterfaceCommand() *cobra.Command {
	var entity string
	var operation string

	cmd := &cobra.Command{
		Use:   "interface [name]",
		Short: "Generate a domain interface",
		Long: `Generate a domain interface (gateway or port) that operates on an existing entity.

This command creates one file in the domain layer:
  1. Interface file declaring I{Name} with a single operation

The operation receives context.Context and a pointer to the entity, and returns error.
The entity struct is assumed to already exist in the domain package. This command does
not check that the entity file is present and does not generate an adapter.

The positional name must be PascalCase and must not include the I prefix.
The I prefix is added automatically (ProductReceiver becomes IProductReceiver).

File names follow the same lowercase convention as other generators
(ProductReceiver is written to productreceiver.go).`,
		Args: cobra.ExactArgs(1),
		Example: `  # Generate a ProductReceiver port for Product
  scaffold-go generate interface ProductReceiver --entity Product --operation ProcessProduct

  # Use the short alias and short entity flag
  scaffold-go gen interface OrderGateway -e Order --operation SubmitOrder`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			projectDir := "."

			logger := GetLogger()
			logger.Info("Starting interface generation",
				zap.String("interface_name", name),
				zap.String("entity", entity),
				zap.String("operation", operation),
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
				return WrapTemplateError("interface", "initialization", err)
			}

			gen := generator.NewInterfaceGenerator(name, entity, operation, projectDir, tmpl)

			logger.Debug("Validating interface parameters")
			if err := gen.Validate(); err != nil {
				logger.Error("Interface validation failed",
					zap.Error(err),
					zap.String("interface_name", name),
				)
				return WrapValidationError("interface", name, err)
			}

			logger.Debug("Generating interface files")
			files, err := gen.Generate()
			if err != nil {
				logger.Error("Interface generation failed",
					zap.Error(err),
					zap.String("interface_name", name),
				)
				return WrapGenerationError("interface", "generation", err)
			}

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

			logger.Info("Interface generated successfully",
				zap.String("interface_name", name),
				zap.Int("files_created", len(files)),
			)
			fmt.Printf("✓ Interface '%s' generated\n", name)
			for _, f := range files {
				fmt.Printf("  - %s\n", f.Path)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&entity, "entity", "e", "", "Existing domain entity the operation receives (PascalCase). Example: Product")
	cmd.Flags().StringVar(&operation, "operation", "", "Operation name on the interface (PascalCase). Example: ProcessProduct")
	_ = cmd.MarkFlagRequired("entity")
	_ = cmd.MarkFlagRequired("operation")

	return cmd
}
