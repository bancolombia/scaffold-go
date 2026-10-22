package commands

import (
	"fmt"
	"strings"

	"github.com/bancolombia/scaffold-go/internal/filesystem"
	"github.com/bancolombia/scaffold-go/internal/generator"
	"github.com/bancolombia/scaffold-go/internal/manager"
	"github.com/bancolombia/scaffold-go/internal/parser"
	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewGenerateUseCaseCommand() *cobra.Command {
	var dependencies []string

	cmd := &cobra.Command{
		Use:   "usecase [name]",
		Short: "Generate a domain use case",
		Long: `Generate a domain use case with interface and implementation.

This command creates a use case file in the domain layer containing:
  • Use case interface definition
  • Use case struct implementation
  • Constructor with dependency injection
  • Execute method for business logic
  • Zap logger integration

Use cases orchestrate business logic and depend only on domain interfaces.
They receive domain entities or primitives as input and return domain entities or primitives.

Dependency format: name:Type
  • name: Variable name (camelCase)
  • Type: Interface type (PascalCase)
  • Example: userRepo:UserRepository,emailService:EmailService`,
		Args: cobra.ExactArgs(1),
		Example: `  # Generate a use case with repository dependency
  scaffold-go generate usecase CreateUser --dependencies userRepo:UserRepository

  # Generate a use case with multiple dependencies
  scaffold-go gen usecase ProcessOrder --dependencies orderRepo:OrderRepository,paymentService:PaymentService

  # Generate a use case without dependencies
  scaffold-go generate usecase ValidateEmail

  # Use short flag notation
  scaffold-go gen usecase GetProduct -d productRepo:ProductRepository`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			projectDir := "."

			logger := GetLogger()
			LogCommandStart(logger, "generate-usecase", name,
				zap.Strings("dependencies", dependencies),
				zap.String("project_dir", projectDir),
			)

			validator := GetValidator()
			LogValidationStart(logger, "project-initialization")
			if err := validator.ValidateProjectInitialized(projectDir); err != nil {
				LogValidationError(logger, "project-initialization", err,
					zap.String("project_dir", projectDir),
				)
				return WrapValidationError("project", projectDir, err)
			}

			logger.Debug("Initializing template engine")
			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				LogTemplateError(logger, "usecase", err)
				return WrapTemplateError("usecase", "initialization", err)
			}

			gen := generator.NewUseCaseGenerator(name, dependencies, projectDir, tmpl)

			LogValidationStart(logger, "usecase-parameters")
			if err := gen.Validate(); err != nil {
				LogValidationError(logger, "usecase-parameters", err,
					zap.String("usecase_name", name),
				)
				return WrapValidationError("usecase", name, err)
			}

			LogGenerationStart(logger, "usecase")
			files, err := gen.Generate()
			if err != nil {
				LogGenerationError(logger, "usecase", err)
				return WrapGenerationError("usecase", "generation", err)
			}
			logger.Debug("Use case files generated",
				zap.Int("file_count", len(files)),
			)

			fs := filesystem.New()
			operations := make([]filesystem.FileOperation, len(files))
			for i, f := range files {
				content := []byte(f.Content)
				if strings.HasSuffix(f.Path, ".go") {
					LogFileOperation(logger, "format", f.Path)
					content, err = fs.FormatGoCode(content)
					if err != nil {
						LogFileOperationError(logger, "format", f.Path, err)
						return WrapFileSystemError("format", f.Path, err)
					}
				}
				operations[i] = filesystem.FileOperation{
					Path:    f.Path,
					Content: content,
				}
			}

			LogFileOperation(logger, "write-batch", projectDir)
			if err := fs.WriteFiles(operations); err != nil {
				LogFileOperationError(logger, "write-batch", projectDir, err)
				return WrapFileSystemError("write", projectDir, err)
			}

			mainPath := "cmd/main.go"
			LogFileOperation(logger, "update", mainPath)
			pathParser := parser.NewModulePathParser(projectDir)
			modulePath, err := pathParser.Parse()
			if err != nil {
				LogFileOperationError(logger, "update", mainPath, err)
				return WrapFileSystemError("update", mainPath, err)
			}
			mainMgr := manager.NewMainFileManager(projectDir)
			if err := mainMgr.AddImport(modulePath + "/internal/usecase"); err != nil {
				LogFileOperationError(logger, "update", mainPath, err)
				return WrapFileSystemError("update", mainPath, err)
			}
			if err := mainMgr.AddFxModule("usecase.Module"); err != nil {
				LogFileOperationError(logger, "update", mainPath, err)
				return WrapFileSystemError("update", mainPath, err)
			}

			filePaths := make([]string, len(files))
			for i, f := range files {
				filePaths[i] = f.Path
			}
			LogCommandSuccess(logger, "generate-usecase", name,
				zap.Int("files_created", len(files)),
				zap.Strings("file_paths", filePaths),
			)
			fmt.Printf("✓ Use case '%s' generated\n", name)
			for _, f := range files {
				fmt.Printf("  - %s\n", f.Path)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&dependencies, "dependencies", "d", []string{}, "Dependencies in format name:Type (comma-separated). Example: userRepo:UserRepository,emailService:EmailService")

	return cmd
}
