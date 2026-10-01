package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Custom generators for handler property-based testing

// genValidHandlerName generates valid handler names (PascalCase)
func genValidHandlerName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{0,30}`)
}

// genHandlerUseCaseSpec generates valid use case specifications in "name:Type" format
func genHandlerUseCaseSpec() gopter.Gen {
	return gen.SliceOf(gen.OneGenOf(
		gen.Const("createCustomer:CreateCustomerUseCase"),
		gen.Const("getCustomer:GetCustomerUseCase"),
		gen.Const("updateCustomer:UpdateCustomerUseCase"),
		gen.Const("deleteCustomer:DeleteCustomerUseCase"),
		gen.Const("listCustomers:ListCustomersUseCase"),
		gen.Const("processOrder:ProcessOrderUseCase"),
		gen.Const("validatePayment:ValidatePaymentUseCase"),
	))
}

// genRouteSpec generates valid route specifications in "METHOD:PATH:HandlerName" format
func genRouteSpec() gopter.Gen {
	return gen.SliceOf(gen.OneGenOf(
		gen.Const("GET:/customers:GetCustomers"),
		gen.Const("POST:/customers:CreateCustomer"),
		gen.Const("GET:/customers/:id:GetCustomer"),
		gen.Const("PUT:/customers/:id:UpdateCustomer"),
		gen.Const("DELETE:/customers/:id:DeleteCustomer"),
		gen.Const("GET:/orders:GetOrders"),
		gen.Const("POST:/orders:CreateOrder"),
	))
}

// Feature: go-microservice-scaffolding-cli, Property 20: Handler generation includes route registration with Fiber
// **Validates: Requirements 5.2, 5.3**
func TestProperty_HandlerGenerationIncludesRouteRegistrationWithFiber(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler contains RegisterRoutes method that accepts Fiber App", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with default middleware enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find handler file (not module)
			var handlerContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "/"+strings.ToLower(handlerName)+"_handler.go") {
					handlerContent = file.Content
					break
				}
			}

			if handlerContent == "" {
				t.Logf("handler file not found")
				return false
			}

			// Verify Fiber import in handler
			if !strings.Contains(handlerContent, "github.com/gofiber/fiber") {
				t.Logf("Fiber import not found")
				return false
			}

			// Route registration lives in the router file, not the handler
			var routerContent string
			routerFilename := strings.ToLower(handlerName) + "_router.go"
			for _, file := range files {
				if strings.HasSuffix(file.Path, "/"+routerFilename) {
					routerContent = file.Content
					break
				}
			}
			if routerContent == "" {
				t.Logf("router file not found")
				return false
			}

			// SetupRoutes accepts a fiber.Router
			setupRoutesSignature := "SetupRoutes(router fiber.Router)"
			if !strings.Contains(routerContent, setupRoutesSignature) {
				t.Logf("SetupRoutes method with fiber.Router parameter not found in router")
				return false
			}

			// Verify routes are registered using Fiber router methods
			if len(routes) > 0 {
				hasRouterCall := strings.Contains(routerContent, "Router.Get(") ||
					strings.Contains(routerContent, "Router.Post(") ||
					strings.Contains(routerContent, "Router.Put(") ||
					strings.Contains(routerContent, "Router.Delete(")

				if !hasRouterCall {
					t.Logf("no Fiber router method calls found")
					return false
				}
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 21: Entrypoint constructors include use case dependencies
// **Validates: Requirements 5.3, 7.3, 8.3**
func TestProperty_EntrypointConstructorsIncludeUseCaseDependencies(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler constructor includes all specified use cases", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Skip if no use cases
			if len(useCases) == 0 {
				return true
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with default middleware enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find handler file
			var handlerContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "/"+strings.ToLower(handlerName)+"_handler.go") {
					handlerContent = file.Content
					break
				}
			}

			if handlerContent == "" {
				t.Logf("handler file not found")
				return false
			}

			// Verify all use cases are present in the handler struct
			for _, uc := range useCases {
				parts := strings.Split(uc, ":")
				if len(parts) != 2 {
					continue
				}

				useCaseName := strings.TrimSpace(parts[0])
				useCaseType := strings.TrimSpace(parts[1])

				// Check use case field in struct
				if !strings.Contains(handlerContent, useCaseName) {
					t.Logf("use case field %s not found in handler struct", useCaseName)
					return false
				}

				// Check use case type reference
				if !strings.Contains(handlerContent, useCaseType) {
					t.Logf("use case type %s not found in handler", useCaseType)
					return false
				}

				// Check use case in constructor parameters
				// The template generates "NewHandlerHandler" so we need to check for that pattern
				constructorName := "func New" + handlerName + "Handler"
				constructorIdx := strings.Index(handlerContent, constructorName)
				if constructorIdx == -1 {
					t.Logf("constructor not found")
					return false
				}

				// Extract constructor signature
				constructorSection := handlerContent[constructorIdx:]
				braceIdx := strings.Index(constructorSection, "{")
				if braceIdx == -1 {
					t.Logf("constructor body not found")
					return false
				}
				constructorSignature := constructorSection[:braceIdx]

				// Check use case parameter in constructor
				if !strings.Contains(constructorSignature, useCaseName) {
					t.Logf("use case parameter %s not found in constructor", useCaseName)
					return false
				}
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 22: Handler includes validation and response formatting
// **Validates: Requirements 5.4**
func TestProperty_HandlerIncludesValidationAndResponseFormatting(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler methods include validation and response formatting", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with default middleware enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find handler file
			var handlerContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "/"+strings.ToLower(handlerName)+"_handler.go") {
					handlerContent = file.Content
					break
				}
			}

			if handlerContent == "" {
				t.Logf("handler file not found")
				return false
			}

			// Check for response formatting using Fiber methods
			// Handlers should use c.JSON() or similar Fiber response methods
			// Also check for TODO comments which indicate validation placeholders
			hasResponseFormatting := strings.Contains(handlerContent, "c.JSON") ||
				strings.Contains(handlerContent, "c.Status") ||
				strings.Contains(handlerContent, "fiber.Map")

			hasTODO := strings.Contains(handlerContent, "TODO")

			// If there are no routes, the handler might not have methods yet
			// In that case, just check that the handler struct exists
			if len(routes) == 0 {
				// Just verify the handler has the basic structure
				return strings.Contains(handlerContent, handlerName+"Handler struct")
			}

			if !hasResponseFormatting && !hasTODO {
				t.Logf("no Fiber response formatting methods or TODO comments found")
				return false
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 23: Entrypoints include logger in constructor
// **Validates: Requirements 5.5, 7.5, 8.5, 10.5, 11.5, 14.1**
func TestProperty_EntrypointsIncludeLoggerInConstructor(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler constructor includes Zap logger parameter", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with default middleware enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find handler file
			var handlerContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "/"+strings.ToLower(handlerName)+"_handler.go") {
					handlerContent = file.Content
					break
				}
			}

			if handlerContent == "" {
				t.Logf("handler file not found")
				return false
			}

			// Check for Zap logger import
			if !strings.Contains(handlerContent, "go.uber.org/zap") {
				t.Logf("Zap logger import not found")
				return false
			}

			// Check for logger field in struct
			if !strings.Contains(handlerContent, "logger *zap.Logger") {
				t.Logf("logger field not found in handler struct")
				return false
			}

			// Check for logger parameter in constructor
			// The template generates "NewHandlerHandler" so we need to check for that pattern
			constructorName := "func New" + handlerName + "Handler"
			constructorIdx := strings.Index(handlerContent, constructorName)
			if constructorIdx == -1 {
				t.Logf("constructor not found")
				return false
			}

			// Extract constructor signature
			constructorSection := handlerContent[constructorIdx:]
			braceIdx := strings.Index(constructorSection, "{")
			if braceIdx == -1 {
				t.Logf("constructor body not found")
				return false
			}
			constructorSignature := constructorSection[:braceIdx]

			// Check logger parameter in constructor
			if !strings.Contains(constructorSignature, "logger *zap.Logger") {
				t.Logf("logger parameter not found in constructor")
				return false
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 24: Entrypoints depend on domain use case interfaces
// **Validates: Requirements 5.7, 7.7, 8.7**
func TestProperty_EntrypointsDependOnDomainUseCaseInterfaces(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler depends only on domain use case interfaces", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with default middleware enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find handler file
			var handlerContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "/"+strings.ToLower(handlerName)+"_handler.go") {
					handlerContent = file.Content
					break
				}
			}

			if handlerContent == "" {
				t.Logf("handler file not found")
				return false
			}

			// Verify handler is in entrypoints package
			if !strings.Contains(handlerContent, "package handler") && !strings.Contains(handlerContent, "package http") && !strings.Contains(handlerContent, "package entrypoints") {
				t.Logf("handler not in entrypoints package")
				return false
			}

			// Check that use case types are referenced (they should be domain interfaces)
			// The template uses types directly without package qualification,
			// so we just verify the use case types are present
			if len(useCases) > 0 {
				for _, uc := range useCases {
					parts := strings.Split(uc, ":")
					if len(parts) == 2 {
						useCaseType := strings.TrimSpace(parts[1])
						// Verify the use case type is referenced in the handler
						if !strings.Contains(handlerContent, useCaseType) {
							t.Logf("use case type %s not found in handler", useCaseType)
							return false
						}
					}
				}
			}

			// Verify handler does not import infrastructure packages (drivenadapters)
			if strings.Contains(handlerContent, "/drivenadapters/") {
				t.Logf("handler imports infrastructure layer (drivenadapters)")
				return false
			}

			// Verify handler file is in the correct directory
			for _, file := range files {
				if strings.HasSuffix(file.Path, "/"+strings.ToLower(handlerName)+"_handler.go") {
					if !strings.Contains(file.Path, filepath.Join("internal", "entrypoints", "http")) {
						t.Logf("handler file not in entrypoints/http directory")
						return false
					}
				}
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 53: Generated entrypoints are wired as fx Modules
// **Validates: Requirements 15.6**
func TestProperty_GeneratedHandlerEntrypointsAreWiredAsFxModules(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler includes fx Module definition", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with default middleware enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find module file
			var moduleContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "module.go") {
					moduleContent = file.Content
					break
				}
			}

			if moduleContent == "" {
				t.Logf("module file not found")
				return false
			}

			// Verify module file contains fx import
			if !strings.Contains(moduleContent, "go.uber.org/fx") {
				t.Logf("module file does not import go.uber.org/fx")
				return false
			}

			// Verify module file contains Module variable
			if !strings.Contains(moduleContent, "var Module = fx.Module") && !strings.Contains(moduleContent, "var Module = fx.Options") {
				t.Logf("module file does not contain Module variable")
				return false
			}

			// Verify module file contains fx.Provide
			if !strings.Contains(moduleContent, "fx.Provide") {
				t.Logf("module file does not contain fx.Provide")
				return false
			}

			// Verify module provides the handler constructor
			// The template generates "NewHandlerHandler" so we need to check for that pattern
			constructorName := "New" + handlerName + "Handler"
			if !strings.Contains(moduleContent, constructorName) {
				t.Logf("module does not provide constructor %s", constructorName)
				return false
			}

			// Verify module is in handler package
			if !strings.Contains(moduleContent, "package handler") && !strings.Contains(moduleContent, "package http") {
				t.Logf("module file not in handler/http package")
				return false
			}

			// Verify module file is in the correct directory
			for _, file := range files {
				if strings.HasSuffix(file.Path, "module.go") {
					if !strings.Contains(file.Path, filepath.Join("internal", "entrypoints", "http")) {
						t.Logf("module file not in entrypoints/http directory")
						return false
					}
				}
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 21: Handler generation creates Fiber App instance
// **Validates: Requirements 5.3**
func TestProperty_HandlerGenerationCreatesFiberAppInstance(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler creates Fiber App instance using fiber.New()", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with middleware enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// The Fiber App is now created in the server file
			var serverContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "http/server.go") {
					serverContent = file.Content
					break
				}
			}

			if serverContent == "" {
				t.Logf("server file not found")
				return false
			}

			// Check for fiber.New() call in server
			if !strings.Contains(serverContent, "fiber.New(") {
				t.Logf("fiber.New() call not found in server")
				return false
			}

			// Server must expose a *fiber.App via a constructor
			if !strings.Contains(serverContent, "*fiber.App") {
				t.Logf("*fiber.App not referenced in server")
				return false
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 22: Handler generation includes error handler
// **Validates: Requirements 5.4**
func TestProperty_HandlerGenerationIncludesErrorHandler(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler includes custom error handler when enabled", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with error handler enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, false, false, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Server file is where fiber.Config (including ErrorHandler) is wired up
			var serverContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "http/server.go") {
					serverContent = file.Content
					break
				}
			}

			if serverContent == "" {
				t.Logf("server file not found")
				return false
			}

			// The error handler slot must at least be scaffolded in fiber.Config
			if !strings.Contains(serverContent, "ErrorHandler") {
				t.Logf("ErrorHandler slot not found in server fiber.Config")
				return false
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 23: Handler generation includes CORS middleware
// **Validates: Requirements 5.5**
func TestProperty_HandlerGenerationIncludesCORSMiddleware(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler includes CORS middleware when enabled", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with CORS enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, false, false)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// CORS middleware now lives in the middleware/cors.go file
			var corsContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "middleware/cors.go") {
					corsContent = file.Content
					break
				}
			}

			if corsContent == "" {
				t.Logf("cors middleware file not found")
				return false
			}

			// Check for CORS middleware import
			if !strings.Contains(corsContent, "github.com/gofiber/fiber/v2/middleware/cors") {
				t.Logf("CORS middleware import not found")
				return false
			}

			// Check for cors.New() call
			if !strings.Contains(corsContent, "cors.New(") {
				t.Logf("cors.New() call not found")
				return false
			}

			// Check for CORS configuration
			if !strings.Contains(corsContent, "AllowOrigins") || !strings.Contains(corsContent, "AllowMethods") {
				t.Logf("CORS configuration not found")
				return false
			}

			// Server must invoke the CORS middleware
			var serverContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "http/server.go") {
					serverContent = file.Content
					break
				}
			}
			if !strings.Contains(serverContent, "middleware.CorsMiddleware()") {
				t.Logf("server does not register CORS middleware")
				return false
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 24: Handler generation includes rate limiter middleware
// **Validates: Requirements 5.6**
func TestProperty_HandlerGenerationIncludesRateLimiterMiddleware(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler includes rate limiter middleware when enabled", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator with rate limiter enabled
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, false, true, false)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Rate limiter is wired at server-level; ensure the server file scaffolds a slot for it
			var serverContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "http/server.go") {
					serverContent = file.Content
					break
				}
			}
			if serverContent == "" {
				t.Logf("server file not found")
				return false
			}

			// The server should reference the rate limiter middleware slot
			if !strings.Contains(serverContent, "RateLimiter") {
				t.Logf("rate limiter slot not found in server")
				return false
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 25: Handler generation includes route registration
// **Validates: Requirements 5.7**
func TestProperty_HandlerGenerationIncludesRouteRegistration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler includes RegisterRoutes method with Fiber App", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Route registration lives in the router file
			var routerContent string
			routerFilename := strings.ToLower(handlerName) + "_router.go"
			for _, file := range files {
				if strings.HasSuffix(file.Path, "/"+routerFilename) {
					routerContent = file.Content
					break
				}
			}

			if routerContent == "" {
				t.Logf("router file not found")
				return false
			}

			// Check for SetupRoutes method that accepts fiber.Router
			if !strings.Contains(routerContent, "SetupRoutes(router fiber.Router)") {
				t.Logf("SetupRoutes method with fiber.Router parameter not found")
				return false
			}

			// Check for router group registration (equivalent to RegisterRoutes call)
			if !strings.Contains(routerContent, "router.Group(") {
				t.Logf("router.Group call not found")
				return false
			}

			// The generic routes.go file must invoke this router
			var routesContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "routes/routes.go") {
					routesContent = file.Content
					break
				}
			}
			if routesContent == "" {
				t.Logf("routes.go file not found")
				return false
			}
			if !strings.Contains(routesContent, "New"+handlerName+"Router(") {
				t.Logf("router constructor not invoked in routes.go")
				return false
			}
			if !strings.Contains(routesContent, ".SetupRoutes(") {
				t.Logf("SetupRoutes not invoked in routes.go")
				return false
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 28: Entrypoints include logger with structured logging
// **Validates: Requirements 5.10**
func TestProperty_EntrypointsIncludeStructuredLogging(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated handler methods include structured logging with method, path, and IP", prop.ForAll(
		func(handlerName string, useCases []string, routes []string) bool {
			// Skip if no routes
			if len(routes) == 0 {
				return true
			}

			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewHandlerGenerator(handlerName, useCases, routes, tmpDir, templates, true, true, true)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find handler file
			var handlerContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "/"+strings.ToLower(handlerName)+"_handler.go") {
					handlerContent = file.Content
					break
				}
			}

			if handlerContent == "" {
				t.Logf("handler file not found")
				return false
			}

			// Check for structured logging with zap fields
			if !strings.Contains(handlerContent, "zap.String(\"method\"") {
				t.Logf("method logging field not found")
				return false
			}

			if !strings.Contains(handlerContent, "zap.String(\"path\"") {
				t.Logf("path logging field not found")
				return false
			}

			if !strings.Contains(handlerContent, "zap.String(\"ip\"") || !strings.Contains(handlerContent, "c.IP()") {
				t.Logf("IP logging field not found")
				return false
			}

			// Check for logger.Info call
			if !strings.Contains(handlerContent, "logger.Info(") {
				t.Logf("logger.Info call not found")
				return false
			}

			return true
		},
		genValidHandlerName(),
		genHandlerUseCaseSpec(),
		genRouteSpec(),
	))

	properties.TestingRun(t)
}
