package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gabheadz/scaffold-go/internal/filesystem"
	"github.com/gabheadz/scaffold-go/internal/generator"
	"github.com/gabheadz/scaffold-go/internal/template"
	"github.com/gabheadz/scaffold-go/internal/validator"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewGenerateDrivenAdapterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "da",
		Short: "Generate a driven adapter",
		Long: `Generate a driven adapter in the DrivenAdapters layer.

Driven adapters are the components the application USES to talk to the outside
world (databases, external APIs, message brokers, caches). You must specify the
kind of driven adapter:

  • database  - Database adapter (PostgreSQL, MongoDB, DynamoDB)
  • api       - External API client
  • publisher - Message queue publisher
  • cache     - Cache adapter (Redis, in-memory)

All driven adapters include:
  • Constructor with dependency injection
  • Zap logger integration
  • Error handling and retry logic
  • uber-go/fx Module for dependency injection`,
		Example: `  # Database driven adapter
  scaffold-go generate da database --type postgres --entity User --fields Name:string,Email:string

  # External API client
  scaffold-go generate da api PaymentClient --url https://api.payment.com

  # Message publisher
  scaffold-go generate da publisher EventPublisher

  # Cache adapter
  scaffold-go generate da cache UserCache`,
	}

	cmd.AddCommand(NewGenerateDaDatabaseCommand())
	cmd.AddCommand(NewGenerateDaApiCommand())
	cmd.AddCommand(NewGenerateDaPublisherCommand())
	cmd.AddCommand(NewGenerateDaCacheCommand())

	return cmd
}

func NewGenerateDaDatabaseCommand() *cobra.Command {
	var (
		dbType     string
		entityName string
		fields     []string
	)

	cmd := &cobra.Command{
		Use:   "database",
		Short: "Generate a database driven adapter",
		Long: `Generate a database adapter implementation for a repository interface.

This command creates a database adapter in the drivenadapters/database layer with:
  • Adapter struct implementing the repository interface
  • Database-specific query implementations (SQL, MongoDB, DynamoDB)
  • CRUD methods: Create, GetByID, Update, Delete, List
  • Constructor with database connection and logger injection
  • Error handling and transaction management
  • uber-go/fx Module for dependency injection

Supported database types:
  • postgres  - PostgreSQL with SQL queries
  • mongodb   - MongoDB with document queries
  • dynamodb  - AWS DynamoDB with key-value operations

Field format: Name:type
  • Used to generate appropriate database queries and mappings
  • Example: Name:string,Email:string,Age:int`,
		Args: cobra.NoArgs,
		Example: `  # Generate a PostgreSQL adapter
  scaffold-go generate da database \
    --type postgres \
    --entity User \
    --fields Name:string,Email:string,Age:int

  # Generate a MongoDB adapter
  scaffold-go gen da database \
    --type mongodb \
    --entity Product \
    --fields Name:string,Price:float64,Stock:int

  # Generate a DynamoDB adapter
  scaffold-go generate da database \
    --type dynamodb \
    --entity Order \
    --fields UserID:string,Total:float64,Status:string

  # Use short flag notation
  scaffold-go gen da database -t postgres -e Customer -f Name:string,Email:string`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := "."

			validator := GetValidator()
			if err := validator.ValidateProjectInitialized(projectDir); err != nil {
				return WrapValidationError("project", projectDir, err)
			}

			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				return WrapTemplateError("da-database", "initialization", err)
			}

			gen := generator.NewDatabaseAdapterGenerator(dbType, entityName, fields, projectDir, tmpl)

			if err := gen.Validate(); err != nil {
				return WrapValidationError("da-database", entityName, err)
			}

			files, err := gen.Generate()
			if err != nil {
				return WrapGenerationError("da-database", "generation", err)
			}

			fs := filesystem.New()
			operations := make([]filesystem.FileOperation, len(files))
			for i, f := range files {
				content := []byte(f.Content)
				if strings.HasSuffix(f.Path, ".go") {
					content, err = fs.FormatGoCode(content)
					if err != nil {
						return WrapFileSystemError("format", f.Path, err)
					}
				}
				operations[i] = filesystem.FileOperation{
					Path:    f.Path,
					Content: content,
				}
			}

			if err := fs.WriteFiles(operations); err != nil {
				return WrapFileSystemError("write", projectDir, err)
			}

			logger := GetLogger()

			projectName := resolveProjectName(validator, projectDir)
			if warn, err := generator.EnsureConfigBlockForDatabase(projectDir, dbType, projectName, tmpl); err != nil {
				return WrapFileSystemError("config-wiring", "config/config.yaml", err)
			} else if warn != nil {
				logger.Warn(warn.Error())
				fmt.Printf("⚠ %s\n", warn.Error())
			}

			logger.Info("Database adapter generated successfully", zap.String("entity", entityName), zap.String("type", dbType))
			fmt.Printf("✓ Database adapter for '%s' generated (%s)\n", entityName, dbType)
			for _, f := range files {
				fmt.Printf("  - %s\n", f.Path)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&dbType, "type", "t", "postgres", "Database type: postgres, mongodb, or dynamodb")
	cmd.Flags().StringVarP(&entityName, "entity", "e", "", "Entity name that this adapter implements the repository for (required)")
	cmd.Flags().StringSliceVarP(&fields, "fields", "f", []string{}, "Entity fields in format Name:type (comma-separated, required). Example: Name:string,Email:string,Age:int")
	cmd.MarkFlagRequired("entity")
	cmd.MarkFlagRequired("fields")

	return cmd
}

func NewGenerateDaApiCommand() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "api [name]",
		Short: "Generate an external API client driven adapter",
		Long: `Generate an external API client adapter.

This command creates an API client in the drivenadapters/api layer with:
  • Client struct with HTTP methods (GET, POST, PUT, DELETE)
  • Request and response handling
  • Constructor with HTTP client and configuration injection
  • Error handling and retry logic
  • Zap logger integration for API call logging
  • uber-go/fx Module for dependency injection

The generated API client implements domain service interfaces and handles external HTTP communication.`,
		Args: cobra.ExactArgs(1),
		Example: `  # Generate an API client with base URL
  scaffold-go generate da api PaymentClient --url https://api.payment.com

  # Generate an API client for notifications
  scaffold-go gen da api NotificationClient --url https://api.notifications.com

  # Generate an API client for third-party service
  scaffold-go generate da api WeatherClient --url https://api.weather.com

  # Use short flag notation
  scaffold-go gen da api ShippingClient -u https://api.shipping.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			projectDir := "."

			validator := GetValidator()
			if err := validator.ValidateProjectInitialized(projectDir); err != nil {
				return WrapValidationError("project", projectDir, err)
			}

			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				return WrapTemplateError("da-api", "initialization", err)
			}

			gen := generator.NewApiClientGenerator(name, baseURL, projectDir, tmpl)

			if err := gen.Validate(); err != nil {
				return WrapValidationError("da-api", name, err)
			}

			files, err := gen.Generate()
			if err != nil {
				return WrapGenerationError("da-api", "generation", err)
			}

			fs := filesystem.New()
			operations := make([]filesystem.FileOperation, len(files))
			for i, f := range files {
				content := []byte(f.Content)
				if strings.HasSuffix(f.Path, ".go") {
					content, err = fs.FormatGoCode(content)
					if err != nil {
						return WrapFileSystemError("format", f.Path, err)
					}
				}
				operations[i] = filesystem.FileOperation{
					Path:    f.Path,
					Content: content,
				}
			}

			if err := fs.WriteFiles(operations); err != nil {
				return WrapFileSystemError("write", projectDir, err)
			}

			logger := GetLogger()
			logger.Info("API client generated successfully", zap.String("name", name))
			fmt.Printf("✓ API client '%s' generated\n", name)
			for _, f := range files {
				fmt.Printf("  - %s\n", f.Path)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&baseURL, "url", "u", "", "Base URL for the external API (e.g., https://api.example.com)")

	return cmd
}

func NewGenerateDaPublisherCommand() *cobra.Command {
	var (
		broker  string
		pattern string
		target  string
	)

	cmd := &cobra.Command{
		Use:   "publisher [name]",
		Short: "Generate a message publisher driven adapter",
		Long: `Generate a message queue publisher adapter.

This command creates a publisher in the drivenadapters/messaging layer with:
  • Publisher struct + Publish/Send/Notify/Request method tailored to the
    chosen broker/pattern
  • Per-component fx module that is auto-wired into cmd/main.go
  • Shared infrastructure module (created on first generation per broker)

Supported brokers (--broker):
  • rabbitmq - based on github.com/bancolombia/reactive-commons-go
  • ibmmq    - based on github.com/bancolombia/commons-mq-go
  • kafka    - placeholder; not implemented yet

Supported patterns (--pattern):
  • rabbitmq: event | command | query | notification
  • ibmmq:    sender | requestreply

The --target flag specifies the destination event name, command name, query
resource, app name, or queue depending on the pattern.`,
		Args: cobra.ExactArgs(1),
		Example: `  # RabbitMQ event publisher
  scaffold-go generate da publisher OrderEventPublisher \
    --broker rabbitmq --pattern event \
    --target order.created

  # RabbitMQ command publisher targeting another service
  scaffold-go gen da publisher SendInvoicePublisher \
    --broker rabbitmq --pattern command \
    --target send-invoice

  # IBM MQ sender
  scaffold-go generate da publisher InvoiceSender \
    --broker ibmmq --pattern sender \
    --target DEV.QUEUE.INVOICES

  # IBM MQ request/reply client
  scaffold-go generate da publisher CatalogQueryClient \
    --broker ibmmq --pattern requestreply \
    --target DEV.QUEUE.CATALOG`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			projectDir := "."

			logger := GetLogger()
			LogCommandStart(logger, "generate-da-publisher", name,
				zap.String("broker", broker),
				zap.String("pattern", pattern),
				zap.String("target", target),
			)

			validator := GetValidator()
			if err := validator.ValidateProjectInitialized(projectDir); err != nil {
				return WrapValidationError("project", projectDir, err)
			}

			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				return WrapTemplateError("da-publisher", "initialization", err)
			}

			gen := generator.NewPublisherGenerator(name, broker, pattern, target, projectDir, tmpl)

			if err := gen.Validate(); err != nil {
				return WrapValidationError("da-publisher", name, err)
			}

			files, err := gen.Generate()
			if err != nil {
				return WrapGenerationError("da-publisher", "generation", err)
			}

			fs := filesystem.New()
			operations := make([]filesystem.FileOperation, len(files))
			for i, f := range files {
				content := []byte(f.Content)
				if strings.HasSuffix(f.Path, ".go") {
					content, err = fs.FormatGoCode(content)
					if err != nil {
						return WrapFileSystemError("format", f.Path, err)
					}
				}
				operations[i] = filesystem.FileOperation{
					Path:    f.Path,
					Content: content,
				}
			}

			if err := fs.WriteFiles(operations); err != nil {
				return WrapFileSystemError("write", projectDir, err)
			}

			modulePath, _ := validator.GetModulePath(projectDir)
			wirePublisherIntoMain(projectDir, modulePath, gen.Broker, name, logger)

			logger.Info("Publisher generated successfully",
				zap.String("name", name),
				zap.String("broker", gen.Broker),
				zap.String("pattern", gen.Pattern),
			)
			fmt.Printf("✓ Publisher '%s' generated (%s/%s)\n", name, gen.Broker, gen.Pattern)
			for _, f := range files {
				fmt.Printf("  - %s\n", f.Path)
			}
			fmt.Println("Run `go mod tidy` to fetch broker dependencies.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&broker, "broker", "b", "", "Messaging broker: rabbitmq | ibmmq | kafka (required)")
	cmd.Flags().StringVarP(&pattern, "pattern", "p", "", "Messaging pattern (required). RabbitMQ: event|command|query|notification. IBM MQ: sender|requestreply")
	cmd.Flags().StringVarP(&target, "target", "t", "", "Destination name (event, command, resource, queue, or app name). Defaults to lowercased component name.")
	_ = cmd.MarkFlagRequired("broker")
	_ = cmd.MarkFlagRequired("pattern")

	return cmd
}

// wirePublisherIntoMain best-effort patches cmd/main.go to import the
// publishers package and the broker foundation, and to add their Module
// references inside the fx.New(...) call.
func wirePublisherIntoMain(projectDir, modulePath, broker, name string, logger *zap.Logger) {
	if modulePath == "" {
		fmt.Println("⚠ could not resolve module path; skipping cmd/main.go auto-wiring")
		return
	}

	infraPkg := generator.FoundationImportInfo(broker)
	if infraPkg.Alias == "" {
		fmt.Printf("⚠ broker %q has no foundation module; skipping cmd/main.go auto-wiring\n", broker)
		return
	}

	tryWireImport(projectDir, modulePath+"/internal/infrastructure/messaging/"+broker, infraPkg.Alias, logger)
	tryWireImport(projectDir, modulePath+"/internal/drivenadapters/messaging", "publishers", logger)

	tryWireModule(projectDir, infraPkg.Alias+".Module", logger)
	tryWireModule(projectDir, "publishers."+name+"Module", logger)
}

func NewGenerateDaCacheCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache [name]",
		Short: "Generate a cache driven adapter",
		Long: `Generate a cache adapter with Get, Set, Delete, and Exists methods.

This command creates a cache adapter in the drivenadapters/cache layer with:
  • Cache struct with standard cache operations
  • Get, Set, Delete, and Exists methods
  • TTL (time-to-live) configuration support
  • Constructor with cache connection injection (Redis or in-memory)
  • Error handling logic
  • uber-go/fx Module for dependency injection

The generated cache adapter provides caching functionality for the application.`,
		Args: cobra.ExactArgs(1),
		Example: `  # Generate a user cache adapter
  scaffold-go generate da cache UserCache

  # Generate a product cache adapter
  scaffold-go gen da cache ProductCache

  # Generate a session cache adapter
  scaffold-go generate da cache SessionCache

  # Generate an API response cache
  scaffold-go gen da cache ApiResponseCache`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			projectDir := "."

			validator := GetValidator()
			if err := validator.ValidateProjectInitialized(projectDir); err != nil {
				return WrapValidationError("project", projectDir, err)
			}

			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				return WrapTemplateError("da-cache", "initialization", err)
			}

			gen := generator.NewCacheGenerator(name, projectDir, tmpl)

			if err := gen.Validate(); err != nil {
				return WrapValidationError("da-cache", name, err)
			}

			files, err := gen.Generate()
			if err != nil {
				return WrapGenerationError("da-cache", "generation", err)
			}

			fs := filesystem.New()
			operations := make([]filesystem.FileOperation, len(files))
			for i, f := range files {
				content := []byte(f.Content)
				if strings.HasSuffix(f.Path, ".go") {
					content, err = fs.FormatGoCode(content)
					if err != nil {
						return WrapFileSystemError("format", f.Path, err)
					}
				}
				operations[i] = filesystem.FileOperation{
					Path:    f.Path,
					Content: content,
				}
			}

			if err := fs.WriteFiles(operations); err != nil {
				return WrapFileSystemError("write", projectDir, err)
			}

			logger := GetLogger()
			logger.Info("Cache adapter generated successfully", zap.String("name", name))
			fmt.Printf("✓ Cache adapter '%s' generated\n", name)
			for _, f := range files {
				fmt.Printf("  - %s\n", f.Path)
			}

			return nil
		},
	}

	return cmd
}

// resolveProjectName returns the last segment of the module path declared in
// go.mod (e.g. "github.com/acme/orders" → "orders"). Falls back to the
// project directory's basename when go.mod cannot be read.
func resolveProjectName(v *validator.Validator, projectDir string) string {
	modulePath, err := v.GetModulePath(projectDir)
	if err == nil && modulePath != "" {
		return filepath.Base(modulePath)
	}
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		return "app"
	}
	return filepath.Base(abs)
}
