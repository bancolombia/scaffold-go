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

func NewGenerateEntryPointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ep",
		Short: "Generate an entry-point",
		Long: `Generate an entry-point in the EntryPoints layer.

Entry-points are the components that DRIVE the application from the outside
(they call the domain use cases). You must specify the kind of entry-point:

  • http      - HTTP handler using the Fiber framework
  • consumer  - Message queue consumer (RabbitMQ, IBM MQ, or Kafka placeholder)

All entry-point components include:
  • Constructor with use case dependency injection
  • Zap logger integration for structured request/message logging
  • uber-go/fx Module for dependency injection`,
		Example: `  # HTTP handler entry-point
  scaffold-go generate ep http UserHandler \
    --usecases createUser:CreateUserUseCase \
    --routes POST:/users:CreateUser

  # Message consumer entry-point (RabbitMQ event)
  scaffold-go generate ep consumer OrderConsumer \
    --broker rabbitmq --pattern event \
    --usecases processOrder:ProcessOrderUseCase`,
	}

	cmd.AddCommand(NewGenerateEpHttpCommand())
	cmd.AddCommand(NewGenerateEpConsumerCommand())

	return cmd
}

func NewGenerateEpHttpCommand() *cobra.Command {
	var (
		useCases           []string
		routes             []string
		enableCORS         bool
		enableRateLimiter  bool
		enableErrorHandler bool
	)

	cmd := &cobra.Command{
		Use:   "http [name]",
		Short: "Generate an HTTP handler entry-point (Fiber framework)",
		Long: `Generate an HTTP handler with Fiber App initialization, middleware, and route registration.

This command creates an HTTP handler in the entrypoints/http layer with:
  • Fiber App instance using fiber.New()
  • Custom error handler for processing Fiber errors
  • CORS middleware configuration (configurable origins, methods, headers)
  • Rate limiter middleware (IP-based limiting with configurable max requests)
  • RegisterRoutes method for route registration
  • GetApp() method to access the Fiber App
  • Handler methods using fiber.Ctx for request/response handling
  • Zap logger integration for structured request logging
  • uber-go/fx Module for dependency injection

Use case format: name:Type
  • name: Variable name (camelCase)
  • Type: Use case interface type (PascalCase)
  • Example: createUser:CreateUserUseCase

Route format: METHOD:path:handler
  • METHOD: HTTP method (GET, POST, PUT, DELETE, PATCH)
  • path: URL path (e.g., /users, /users/:id)
  • handler: Handler method name (PascalCase)
  • Example: POST:/users:CreateUser,GET:/users/:id:GetUser`,
		Args: cobra.ExactArgs(1),
		Example: `  # Generate a handler with all middleware enabled
  scaffold-go generate ep http UserHandler \
    --usecases createUser:CreateUserUseCase,getUser:GetUserUseCase \
    --routes POST:/users:CreateUser,GET:/users/:id:GetUser \
    --enable-cors --enable-rate-limiter --enable-error-handler

  # Generate a simple handler with CORS only
  scaffold-go gen ep http ProductHandler \
    --usecases getProduct:GetProductUseCase \
    --routes GET:/products/:id:GetProduct \
    --enable-cors

  # Generate a handler without middleware
  scaffold-go generate ep http HealthHandler \
    --routes GET:/health:CheckHealth \
    --enable-cors=false --enable-rate-limiter=false --enable-error-handler=false

  # Use short flag notation
  scaffold-go gen ep http OrderHandler \
    -u processOrder:ProcessOrderUseCase \
    -r POST:/orders:CreateOrder,GET:/orders/:id:GetOrder`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			projectDir := "."

			logger := GetLogger()
			LogCommandStart(logger, "generate-ep-http", name,
				zap.Strings("usecases", useCases),
				zap.Strings("routes", routes),
				zap.Bool("enable_cors", enableCORS),
				zap.Bool("enable_rate_limiter", enableRateLimiter),
				zap.Bool("enable_error_handler", enableErrorHandler),
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
				LogTemplateError(logger, "ep-http", err)
				return WrapTemplateError("ep-http", "initialization", err)
			}

			// if name does end with "Handler", remove it
			if before, ok := strings.CutSuffix(name, "Handler"); ok {
				name = before
			}

			gen := generator.NewHandlerGenerator(name, useCases, routes, projectDir, tmpl, enableCORS, enableRateLimiter, enableErrorHandler)

			LogValidationStart(logger, "ep-http-parameters")
			if err := gen.Validate(); err != nil {
				LogValidationError(logger, "ep-http-parameters", err,
					zap.String("handler_name", name),
				)
				return WrapValidationError("ep-http", name, err)
			}

			LogGenerationStart(logger, "ep-http")
			files, err := gen.Generate()
			if err != nil {
				LogGenerationError(logger, "ep-http", err)
				return WrapGenerationError("ep-http", "generation", err)
			}
			logger.Debug("HTTP handler files generated",
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

			filePaths := make([]string, len(files))
			for i, f := range files {
				filePaths[i] = f.Path
			}
			LogCommandSuccess(logger, "generate-ep-http", name,
				zap.Int("files_created", len(files)),
				zap.Strings("file_paths", filePaths),
			)
			fmt.Printf("✓ HTTP handler '%s' generated\n", name)
			for _, f := range files {
				fmt.Printf("  - %s\n", f.Path)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&useCases, "usecases", "u", []string{}, "Use cases in format name:Type (comma-separated). Example: createUser:CreateUserUseCase,getUser:GetUserUseCase")
	cmd.Flags().StringSliceVarP(&routes, "routes", "r", []string{}, "Routes in format METHOD:path:handler (comma-separated). Example: POST:/users:CreateUser,GET:/users/:id:GetUser")
	cmd.Flags().BoolVar(&enableCORS, "enable-cors", true, "Enable CORS middleware with configurable origins, methods, and headers")
	cmd.Flags().BoolVar(&enableRateLimiter, "enable-rate-limiter", true, "Enable rate limiter middleware with IP-based limiting")
	cmd.Flags().BoolVar(&enableErrorHandler, "enable-error-handler", true, "Enable custom error handler for processing Fiber errors")

	return cmd
}

func NewGenerateEpConsumerCommand() *cobra.Command {
	var (
		broker   string
		pattern  string
		useCases []string
		queue    string
	)

	cmd := &cobra.Command{
		Use:   "consumer [name]",
		Short: "Generate a message consumer entry-point",
		Long: `Generate a message queue consumer with use case dependencies.

This command creates a consumer in the entrypoints/messaging layer with:
  • Consumer struct + Handle method tailored to the chosen broker/pattern
  • fx.Lifecycle registration so the consumer starts with the application
  • Per-component fx module that is auto-wired into cmd/main.go
  • Shared infrastructure module (created on first generation per broker)

Supported brokers (--broker):
  • rabbitmq - based on github.com/bancolombia/reactive-commons-go
  • ibmmq    - based on github.com/bancolombia/commons-mq-go
  • kafka    - placeholder; not implemented yet

Supported patterns (--pattern):
  • rabbitmq: event | command | query | notification
  • ibmmq:    listener | requestreply

Use case format: name:Type[:Package]
  • name:    Variable name (camelCase)
  • Type:    Use case interface type (PascalCase)
  • Package: Optional package alias for qualified imports
  • Example: processOrder:ProcessOrderUseCase`,
		Args: cobra.ExactArgs(1),
		Example: `  # RabbitMQ event consumer
  scaffold-go generate ep consumer OrderEventConsumer \
    --broker rabbitmq --pattern event \
    --usecases processOrder:ProcessOrderUseCase \
    --queue order.created

  # IBM MQ listener
  scaffold-go gen ep consumer InvoiceListener \
    --broker ibmmq --pattern listener \
    --queue DEV.QUEUE.INVOICES

  # RabbitMQ async query handler
  scaffold-go generate ep consumer ProductQueryHandler \
    --broker rabbitmq --pattern query \
    --usecases getProduct:GetProductUseCase \
    --queue get-product`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			projectDir := "."

			logger := GetLogger()
			LogCommandStart(logger, "generate-ep-consumer", name,
				zap.String("broker", broker),
				zap.String("pattern", pattern),
				zap.Strings("usecases", useCases),
				zap.String("queue", queue),
			)

			validator := GetValidator()
			if err := validator.ValidateProjectInitialized(projectDir); err != nil {
				return WrapValidationError("project", projectDir, err)
			}

			tmpl, err := template.NewTemplateEngine()
			if err != nil {
				return WrapTemplateError("ep-consumer", "initialization", err)
			}

			gen := generator.NewConsumerGenerator(name, broker, pattern, useCases, queue, projectDir, tmpl)

			if err := gen.Validate(); err != nil {
				return WrapValidationError("ep-consumer", name, err)
			}

			files, err := gen.Generate()
			if err != nil {
				return WrapGenerationError("ep-consumer", "generation", err)
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
			wireConsumerIntoMain(projectDir, modulePath, gen.Broker, name, logger)

			logger.Info("Consumer generated successfully",
				zap.String("name", name),
				zap.String("broker", gen.Broker),
				zap.String("pattern", gen.Pattern),
			)
			fmt.Printf("✓ Consumer '%s' generated (%s/%s)\n", name, gen.Broker, gen.Pattern)
			for _, f := range files {
				fmt.Printf("  - %s\n", f.Path)
			}
			fmt.Println("Run `go mod tidy` to fetch broker dependencies.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&broker, "broker", "b", "", "Messaging broker: rabbitmq | ibmmq | kafka (required)")
	cmd.Flags().StringVarP(&pattern, "pattern", "p", "", "Messaging pattern (required). RabbitMQ: event|command|query|notification. IBM MQ: listener|requestreply")
	cmd.Flags().StringSliceVarP(&useCases, "usecases", "u", []string{}, "Use cases in format name:Type[:Package] (comma-separated). Example: processOrder:ProcessOrderUseCase")
	cmd.Flags().StringVarP(&queue, "queue", "q", "", "Queue, topic, event, or resource name. Defaults to lowercased component name.")
	_ = cmd.MarkFlagRequired("broker")
	_ = cmd.MarkFlagRequired("pattern")

	return cmd
}

// wireConsumerIntoMain best-effort patches cmd/main.go to import the consumers
// package and the broker foundation, and to add their Module references inside
// the fx.New(...) call. Failures emit a warning but do not abort the
// generation, since the user can always wire things up by hand.
func wireConsumerIntoMain(projectDir, modulePath, broker, name string, logger *zap.Logger) {
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
	tryWireImport(projectDir, modulePath+"/internal/entrypoints/messaging", "consumers", logger)

	tryWireModule(projectDir, infraPkg.Alias+".Module", logger)
	tryWireModule(projectDir, "consumers."+name+"Module", logger)
}

func tryWireImport(projectDir, importPath, alias string, logger *zap.Logger) {
	if err := generator.EnsureMainImport(projectDir, importPath, alias); err != nil {
		var w *generator.MainWiringWarning
		if asMainWiringWarning(err, &w) {
			fmt.Printf("⚠ %s\n", w.Error())
		} else {
			logger.Warn("auto-wire import failed", zap.Error(err))
		}
	}
}

func tryWireModule(projectDir, expr string, logger *zap.Logger) {
	if err := generator.EnsureMainModule(projectDir, expr); err != nil {
		var w *generator.MainWiringWarning
		if asMainWiringWarning(err, &w) {
			fmt.Printf("⚠ %s\n", w.Error())
		} else {
			logger.Warn("auto-wire module failed", zap.Error(err))
		}
	}
}
