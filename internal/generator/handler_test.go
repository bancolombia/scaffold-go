package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerGenerator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		handlerName string
		wantErr     bool
	}{
		{
			name:        "valid handler name",
			handlerName: "Customer",
			wantErr:     false,
		},
		{
			name:        "invalid handler name with special chars",
			handlerName: "Customer-Handler!",
			wantErr:     true,
		},
		{
			name:        "invalid handler name starting with number",
			handlerName: "123Handler",
			wantErr:     true,
		},
		{
			name:        "empty handler name",
			handlerName: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			gen := NewHandlerGenerator(tt.handlerName, nil, nil, "", templates, true, true, true)
			err = gen.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandlerGenerator_Generate(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	useCases := []string{"createCustomer:CreateCustomerUseCase", "getCustomer:GetCustomerUseCase"}
	routes := []string{"POST:/customers:CreateCustomer", "GET:/customers/:id:GetCustomer"}

	gen := NewHandlerGenerator("Customer", useCases, routes, tempDir, templates, true, true, true)

	files, err := gen.Generate()
	require.NoError(t, err)

	byPath := indexFilesByPath(files)

	// Handler file
	handlerPath := filepath.Join(tempDir, "internal", "entrypoints", "http", "handler", "customer_handler.go")
	handlerContent, ok := byPath[handlerPath]
	require.True(t, ok, "handler file not generated at %s", handlerPath)
	assert.Contains(t, handlerContent, "type CustomerHandler struct")
	assert.Contains(t, handlerContent, "createCustomer usecase.CreateCustomerUseCase")
	assert.Contains(t, handlerContent, "getCustomer usecase.GetCustomerUseCase")
	assert.Contains(t, handlerContent, "func NewCustomerHandler")
	assert.Contains(t, handlerContent, "logger *zap.Logger")
	assert.Contains(t, handlerContent, "func (h *CustomerHandler) CreateCustomer(c *fiber.Ctx) error")
	assert.Contains(t, handlerContent, "func (h *CustomerHandler) GetCustomer(c *fiber.Ctx) error")

	// Router file (route registration lives here now)
	routerPath := filepath.Join(tempDir, "internal", "entrypoints", "http", "routes", "customer_router.go")
	routerContent, ok := byPath[routerPath]
	require.True(t, ok, "router file not generated at %s", routerPath)
	assert.Contains(t, routerContent, "SetupRoutes(router fiber.Router)")
	assert.Contains(t, routerContent, "customerRouter.Post(")
	assert.Contains(t, routerContent, "customerRouter.Get(")

	// Handler module file
	modulePath := filepath.Join(tempDir, "internal", "entrypoints", "http", "handler", "module.go")
	moduleContent, ok := byPath[modulePath]
	require.True(t, ok, "module file not generated at %s", modulePath)
	assert.Contains(t, moduleContent, "package handler")
	assert.Contains(t, moduleContent, "fx.Module")
	assert.Contains(t, moduleContent, "NewCustomerHandler")
}

func indexFilesByPath(files []GeneratedFile) map[string]string {
	byPath := make(map[string]string, len(files))
	for _, f := range files {
		byPath[f.Path] = f.Content
	}
	return byPath
}

func TestHandlerGenerator_ComponentType(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewHandlerGenerator("Customer", nil, nil, "", templates, true, true, true)
	assert.Equal(t, "handler", gen.ComponentType())
}

func TestHandlerGenerator_ParseUseCases(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	useCases := []string{
		"createCustomer:usecase.CreateCustomerUseCase",
		"getCustomer:usecase.GetCustomerUseCase",
	}

	gen := NewHandlerGenerator("Customer", useCases, nil, "", templates, true, true, true)
	parsed := gen.parseUseCases()

	assert.Len(t, parsed, 2)
	assert.Equal(t, "createCustomer", parsed[0].Name)
	assert.Equal(t, "usecase.CreateCustomerUseCase", parsed[0].Type)
	assert.Equal(t, "getCustomer", parsed[1].Name)
	assert.Equal(t, "usecase.GetCustomerUseCase", parsed[1].Type)
}

func TestHandlerGenerator_ParseRoutes(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	routes := []string{
		"POST:/customers:CreateCustomer",
		"GET:/customers/:id:GetCustomer",
		"put:/customers/:id:UpdateCustomer",
	}

	gen := NewHandlerGenerator("Customer", nil, routes, "", templates, true, true, true)
	parsed := gen.parseRoutes()

	assert.Len(t, parsed, 3)

	assert.Equal(t, "Post", parsed[0].Method)
	assert.Equal(t, "/customers", parsed[0].Path)
	assert.Equal(t, "CreateCustomer", parsed[0].Handler)

	assert.Equal(t, "Get", parsed[1].Method)
	assert.Equal(t, "/customers/:id", parsed[1].Path)
	assert.Equal(t, "GetCustomer", parsed[1].Handler)

	assert.Equal(t, "Put", parsed[2].Method)
	assert.Equal(t, "/customers/:id", parsed[2].Path)
	assert.Equal(t, "UpdateCustomer", parsed[2].Handler)
}

func TestHandlerGenerator_Integration(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	useCases := []string{"createCustomer:usecase.CreateCustomerUseCase"}
	routes := []string{"POST:/customers:CreateCustomer"}

	gen := NewHandlerGenerator("Customer", useCases, routes, tempDir, templates, true, true, true)

	// Validate
	err = gen.Validate()
	require.NoError(t, err)

	// Generate
	files, err := gen.Generate()
	require.NoError(t, err)

	// Write files
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(file.Path, []byte(file.Content), 0644)
		require.NoError(t, err)
	}

	// Verify files exist
	handlerPath := filepath.Join(tempDir, "internal", "entrypoints", "http", "handler", "customer_handler.go")
	modulePath := filepath.Join(tempDir, "internal", "entrypoints", "http", "handler", "module.go")

	assert.FileExists(t, handlerPath)
	assert.FileExists(t, modulePath)

	// Read and verify handler content
	handlerContent, err := os.ReadFile(handlerPath)
	require.NoError(t, err)
	assert.Contains(t, string(handlerContent), "package handler")
	assert.Contains(t, string(handlerContent), "fiber.Ctx")

	// Read and verify module content
	moduleContent, err := os.ReadFile(modulePath)
	require.NoError(t, err)
	assert.Contains(t, string(moduleContent), "package handler")
	assert.Contains(t, string(moduleContent), "fx.Module")
}

func TestHandlerGenerator_EmptyUseCasesAndRoutes(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewHandlerGenerator("Customer", []string{}, []string{}, tempDir, templates, true, true, true)

	files, err := gen.Generate()
	require.NoError(t, err)

	byPath := indexFilesByPath(files)
	handlerContent, ok := byPath[filepath.Join(tempDir, "internal", "entrypoints", "http", "handler", "customer_handler.go")]
	require.True(t, ok, "customer_handler.go not generated")
	assert.Contains(t, handlerContent, "type CustomerHandler struct")
	assert.Contains(t, handlerContent, "logger *zap.Logger")
}

func TestHandlerGenerator_MultipleRoutesSameMethod(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	routes := []string{
		"GET:/customers:ListCustomers",
		"GET:/customers/:id:GetCustomer",
		"GET:/customers/:id/orders:GetCustomerOrders",
	}

	gen := NewHandlerGenerator("Customer", nil, routes, tempDir, templates, true, true, true)

	files, err := gen.Generate()
	require.NoError(t, err)

	byPath := indexFilesByPath(files)
	handlerContent, ok := byPath[filepath.Join(tempDir, "internal", "entrypoints", "http", "handler", "customer_handler.go")]
	require.True(t, ok, "customer_handler.go not generated")
	assert.Contains(t, handlerContent, "func (h *CustomerHandler) ListCustomers(c *fiber.Ctx) error")
	assert.Contains(t, handlerContent, "func (h *CustomerHandler) GetCustomer(c *fiber.Ctx) error")
	assert.Contains(t, handlerContent, "func (h *CustomerHandler) GetCustomerOrders(c *fiber.Ctx) error")
}

func TestHandlerGenerator_CaseInsensitiveHTTPMethods(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	routes := []string{
		"get:/customers:GetCustomer",
		"post:/customers:CreateCustomer",
		"Put:/customers/:id:UpdateCustomer",
		"DELETE:/customers/:id:DeleteCustomer",
	}

	gen := NewHandlerGenerator("Customer", nil, routes, tempDir, templates, true, true, true)

	files, err := gen.Generate()
	require.NoError(t, err)

	byPath := indexFilesByPath(files)
	// Route registration now lives in the router file
	routerContent, ok := byPath[filepath.Join(tempDir, "internal", "entrypoints", "http", "routes", "customer_router.go")]
	require.True(t, ok, "customer_router.go not generated")
	// All methods should be Title-cased when routing
	assert.Contains(t, routerContent, "customerRouter.Get(")
	assert.Contains(t, routerContent, "customerRouter.Post(")
	assert.Contains(t, routerContent, "customerRouter.Put(")
	assert.Contains(t, routerContent, "customerRouter.Delete(")
}

func TestHandlerGenerator_LoggerIntegration(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	routes := []string{"GET:/customers:GetCustomer"}

	gen := NewHandlerGenerator("Customer", nil, routes, tempDir, templates, true, true, true)

	files, err := gen.Generate()
	require.NoError(t, err)

	byPath := indexFilesByPath(files)
	handlerContent, ok := byPath[filepath.Join(tempDir, "internal", "entrypoints", "http", "handler", "customer_handler.go")]
	require.True(t, ok, "customer_handler.go not generated")

	// Check logger is in struct and constructor
	assert.Contains(t, handlerContent, "logger *zap.Logger")
	assert.Contains(t, handlerContent, "logger: logger")

	// Check logger is used in handler methods
	assert.Contains(t, handlerContent, "h.logger.Info")
	assert.Contains(t, handlerContent, `zap.String("path"`)
	assert.Contains(t, handlerContent, `zap.String("method"`)
}

func TestHandlerGenerator_FiberFramework(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	routes := []string{"GET:/customers:GetCustomer"}

	gen := NewHandlerGenerator("Customer", nil, routes, tempDir, templates, true, true, true)

	files, err := gen.Generate()
	require.NoError(t, err)

	byPath := indexFilesByPath(files)
	handlerContent, ok := byPath[filepath.Join(tempDir, "internal", "entrypoints", "http", "handler", "customer_handler.go")]
	require.True(t, ok, "customer_handler.go not generated")

	// Check Fiber imports
	assert.Contains(t, handlerContent, `"github.com/gofiber/fiber/v2"`)

	// Check Fiber types are used in handler methods
	assert.Contains(t, handlerContent, "fiber.Ctx")
	assert.Contains(t, handlerContent, "fiber.Map")

	// Check handler method signature
	assert.Contains(t, handlerContent, "func (h *CustomerHandler) GetCustomer(c *fiber.Ctx) error")

	// The Fiber App instance is created in server.go
	serverContent, ok := byPath[filepath.Join(tempDir, "internal", "entrypoints", "http", "server.go")]
	require.True(t, ok, "server.go not generated")
	assert.Contains(t, serverContent, "fiber.New(")
	assert.Contains(t, serverContent, "*fiber.App")

	// Route registration lives in the router
	routerContent, ok := byPath[filepath.Join(tempDir, "internal", "entrypoints", "http", "routes", "customer_router.go")]
	require.True(t, ok, "customer_router.go not generated")
	assert.Contains(t, routerContent, "SetupRoutes(router fiber.Router)")
}

func TestHandlerGenerator_FxModuleStructure(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewHandlerGenerator("Customer", nil, nil, tempDir, templates, true, true, true)

	files, err := gen.Generate()
	require.NoError(t, err)

	byPath := indexFilesByPath(files)
	moduleContent, ok := byPath[filepath.Join(tempDir, "internal", "entrypoints", "http", "handler", "module.go")]
	require.True(t, ok, "handler/module.go not generated")

	// Check fx module structure
	assert.Contains(t, moduleContent, `package handler`)
	assert.Contains(t, moduleContent, `import "go.uber.org/fx"`)
	assert.Contains(t, moduleContent, `var Module = fx.Module(`)
	assert.Contains(t, moduleContent, `fx.Provide(`)
	assert.Contains(t, moduleContent, `NewCustomerHandler`)
}

func TestHandlerGenerator_RouteRegistration(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	routes := []string{
		"GET:/api/v1/customers:ListCustomers",
		"POST:/api/v1/customers:CreateCustomer",
	}

	gen := NewHandlerGenerator("Customer", nil, routes, tempDir, templates, true, true, true)

	files, err := gen.Generate()
	require.NoError(t, err)

	byPath := indexFilesByPath(files)
	routerContent, ok := byPath[filepath.Join(tempDir, "internal", "entrypoints", "http", "routes", "customer_router.go")]
	require.True(t, ok, "customer_router.go not generated")

	// Route registration is done in the router file
	assert.Contains(t, routerContent, "func (r *CustomerRouter) SetupRoutes(router fiber.Router)")
	assert.Contains(t, routerContent, "customerRouter.Get(")
	assert.Contains(t, routerContent, "customerRouter.Post(")
	assert.Contains(t, routerContent, "r.handler.ListCustomers")
	assert.Contains(t, routerContent, "r.handler.CreateCustomer")
}
