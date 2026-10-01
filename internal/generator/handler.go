package generator

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/bancolombia/scaffold-go/internal/manager"
	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/bancolombia/scaffold-go/internal/validator"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type HandlerGenerator struct {
	Name               string
	UseCases           []string
	Routes             []string
	ProjectDir         string
	Templates          *template.TemplateEngine
	Validator          *validator.Validator
	EnableCORS         bool
	EnableRateLimiter  bool
	EnableErrorHandler bool
}

func NewHandlerGenerator(name string, useCases []string, routes []string, projectDir string, templates *template.TemplateEngine, enableCORS, enableRateLimiter, enableErrorHandler bool) *HandlerGenerator {
	return &HandlerGenerator{
		Name:               name,
		UseCases:           useCases,
		Routes:             routes,
		ProjectDir:         projectDir,
		Templates:          templates,
		Validator:          validator.New(),
		EnableCORS:         enableCORS,
		EnableRateLimiter:  enableRateLimiter,
		EnableErrorHandler: enableErrorHandler,
	}
}

func (g *HandlerGenerator) Validate() error {
	if err := g.Validator.ValidateName(g.Name); err != nil {
		return fmt.Errorf("invalid handler name: %w", err)
	}
	return nil
}

func (g *HandlerGenerator) Generate() ([]GeneratedFile, error) {
	useCases := g.parseUseCases()
	routes := g.parseRoutes()

	// Get module path from go.mod
	mp, err := g.Validator.GetModulePath(g.ProjectDir)
	if err != nil {
		// If we can't get module path, use empty string (tests might not have go.mod)
		mp = ""
	}

	data := template.HandlerTemplateData{
		PackageName:        "http",
		HandlerName:        g.Name,
		UseCases:           useCases,
		Routes:             routes,
		HasValidation:      true,
		ModulePath:         mp,
		EnableCORS:         g.EnableCORS,
		EnableRateLimiter:  g.EnableRateLimiter,
		EnableErrorHandler: g.EnableErrorHandler,
	}

	generatedFiles := []GeneratedFile{}

	// ------------------------------
	// ----- Generate Handlers ------
	// ------------------------------
	defaultHealthHandlerPath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "http", "handler", "health_handler.go")
	// check if the health handler already exists
	if err := g.Validator.FileExists(defaultHealthHandlerPath); err == nil {
		// Health handler already exists, skip generation
		log.Println("Health handler already exists, skipping generation")
	} else {
		// Generate default health handler file
		healthHandlerContent, err := g.Templates.Execute("ep/http/handler/health_handler.go.tmpl", data)
		generatedFiles = append(generatedFiles, GeneratedFile{Path: defaultHealthHandlerPath, Content: healthHandlerContent})
		if err != nil {
			return nil, fmt.Errorf("failed to generate health handler: %w", err)
		}
	}

	// Generate the requested new handler file
	handlerContent, err := g.Templates.Execute("ep/http/handler/http_handler.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate handler: %w", err)
	}
	handlerPath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "http", "handler", strings.ToLower(g.Name)+"_handler.go")
	generatedFiles = append(generatedFiles, GeneratedFile{Path: handlerPath, Content: handlerContent})

	// ------------------------------
	// ----- Generate Routes --------
	// ------------------------------

	// First, the generic routes file. Check if already exists
	defaultRoutesPath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "http", "routes", "routes.go")
	if err := g.Validator.FileExists(defaultRoutesPath); err == nil {
		// Generic routes file already exists, skip generation
		// TODO: Consider merging new routes into the existing file instead of skipping generation
		log.Println("Generic routes file already exists, skipping generation")
	} else {
		// Generate the generic routes file
		routesGenricContent, err := g.Templates.Execute("ep/http/routes/routes.go.tmpl", data)
		if err != nil {
			return nil, fmt.Errorf("failed to generate routes: %w", err)
		}
		generatedFiles = append(generatedFiles, GeneratedFile{Path: defaultRoutesPath, Content: routesGenricContent})

		// Generate router for health endpoint
		healthRouteContent, err := g.Templates.Execute("ep/http/routes/health_router.go.tmpl", data)
		if err != nil {
			return nil, fmt.Errorf("failed to generate health route: %w", err)
		}
		healthRouterPath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "http", "routes", "health_router.go")
		generatedFiles = append(generatedFiles, GeneratedFile{Path: healthRouterPath, Content: healthRouteContent})
	}

	// Generate specific router for the new handler
	newRouteContent, err := g.Templates.Execute("ep/http/routes/router.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate specific route: %w", err)
	}
	newRoutePath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "http", "routes", strings.ToLower(g.Name)+"_router.go")
	generatedFiles = append(generatedFiles, GeneratedFile{Path: newRoutePath, Content: newRouteContent})

	// --------------------------------
	// ----- Generate Middlewares -----
	// --------------------------------

	corsPath, corsContent, err := g.generateCorsMiddleware(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cors middleware: %w", err)
	}
	if corsContent != "" {
		// collect content for the generated CORS middleware
		generatedFiles = append(generatedFiles, GeneratedFile{Path: corsPath, Content: corsContent})
	}

	// ---------------------------------
	// ----- Generate Fiber Server -----
	// ---------------------------------

	serverPath, serverContent, err := g.generateFiberServer(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate fiber server: %w", err)
	}
	if serverContent != "" {
		// collect content for the generated Fiber server
		generatedFiles = append(generatedFiles, GeneratedFile{Path: serverPath, Content: serverContent})
	}

	// -------------------------------------
	// ----- Generate/Modify Fx Module -----
	// -------------------------------------
	modulePath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "http", "handler", "module.go")
	if err := g.Validator.FileExists(modulePath); err == nil {
		// TODO: Handle the case where the HTTP handler module file already exists
	} else {
		moduleData := template.HandlerTemplateData{
			PackageName: "handler",
			HandlerName: g.Name,
		}
		moduleContent, err := g.Templates.Execute("ep/http/handler/module.go.tmpl", moduleData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate module: %w", err)
		}

		generatedFiles = append(generatedFiles, GeneratedFile{Path: modulePath, Content: moduleContent})
	}

	// --------------------------------------
	// ----- Add HTTP Module to main.go -----
	// --------------------------------------
	mainMgr := manager.NewMainFileManager(g.ProjectDir)
	err = mainMgr.AddImport(mp + "/internal/entrypoints/http")
	if err != nil {
		log.Printf("Result of adding http import to main.go: %s", err)
	}
	err = mainMgr.AddFxModule("http.Module")
	if err != nil {
		log.Printf("Result of adding http module to main.go: %s", err)
	}

	return generatedFiles, nil
}

func (g *HandlerGenerator) generateCorsMiddleware(data template.HandlerTemplateData) (string, string, error) {
	middlewarePath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "http", "middleware", "cors.go")
	if err := g.Validator.FileExists(middlewarePath); err == nil {
		// cors middleware already exists
		log.Println("CORS middleware already exists, skipping generation")
	} else {
		middlewareContent, err := g.Templates.Execute("ep/http/middleware/cors.go.tmpl", data)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate cors middleware: %w", err)
		}
		return middlewarePath, middlewareContent, nil
	}
	return "", "", nil
}

func (g *HandlerGenerator) generateFiberServer(data template.HandlerTemplateData) (string, string, error) {
	serverPath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "http", "server.go")
	if err := g.Validator.FileExists(serverPath); err == nil {
		// fiber server already exists
		log.Println("Fiber server already exists, skipping generation")
	} else {
		serverContent, err := g.Templates.Execute("ep/http/server.go.tmpl", data)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate fiber server: %w", err)
		}
		return serverPath, serverContent, nil
	}
	return "", "", nil
}

func (g *HandlerGenerator) ComponentType() string {
	return "handler"
}

func (g *HandlerGenerator) parseUseCases() []template.Dependency {
	useCases := make([]template.Dependency, 0, len(g.UseCases))
	for _, uc := range g.UseCases {
		parts := strings.Split(uc, ":")
		if len(parts) == 2 {
			useCases = append(useCases, template.Dependency{
				Name: strings.TrimSpace(parts[0]),
				Type: strings.TrimSpace(parts[1]),
			})
		}
	}
	return useCases
}

func (g *HandlerGenerator) parseRoutes() []template.Route {
	routes := make([]template.Route, 0, len(g.Routes))
	for _, r := range g.Routes {
		// Parse route format: "METHOD:PATH:HandlerName"
		// We need to be careful with colons in the path (e.g., /customers/:id)
		// Split only on the first two colons
		firstColon := strings.Index(r, ":")
		if firstColon == -1 {
			continue
		}

		methodRaw := strings.TrimSpace(r[:firstColon])
		// Capitalize only the first letter (e.g., "get" -> "Get", "GET" -> "Get")
		//method := strings.Title(strings.ToLower(methodRaw))
		method := cases.Title(language.Und, cases.NoLower).String(strings.ToLower(methodRaw))
		remainder := r[firstColon+1:]

		// Find the last colon for the handler name
		lastColon := strings.LastIndex(remainder, ":")
		if lastColon == -1 {
			continue
		}

		path := strings.TrimSpace(remainder[:lastColon])
		handler := strings.TrimSpace(remainder[lastColon+1:])

		routes = append(routes, template.Route{
			Method:  method,
			Path:    path,
			Handler: handler,
		})
	}
	return routes
}
