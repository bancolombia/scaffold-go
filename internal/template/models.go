package template

// Field represents an entity field
type Field struct {
	Name     string   // Field name (PascalCase)
	Type     string   // Go type (string, int, time.Time, etc.)
	Tags     []string // Struct tags (json, db, validate)
	Required bool     // Whether field is required
}

// Route represents an HTTP route
type Route struct {
	Method  string // HTTP method (GET, POST, PUT, DELETE)
	Path    string // URL path
	Handler string // Handler function name
}

// Dependency represents an injected dependency
type Dependency struct {
	Name       string // Variable name
	Type       string // Interface type
	Package    string // Import path
	EntityType string // Entity name extracted from repository type (e.g., "User" from "UserRepository")
}

// Method represents an interface method
type Method struct {
	Name       string
	Parameters []Parameter
	Returns    []string
}

// Parameter represents a method parameter
type Parameter struct {
	Name string
	Type string
}

// EntityTemplateData contains data for entity template
type EntityTemplateData struct {
	PackageName    string
	EntityName     string
	Fields         []Field
	HasTimeFields  bool
	HasValidation  bool
	RepositoryName string
	ModulePath     string
}

// DomainInterfaceTemplateData contains data for a domain gateway/port template.
type DomainInterfaceTemplateData struct {
	PackageName   string
	InterfaceName string
	EntityName    string
	OperationName string
	ParamName     string
}

// UseCaseTemplateData contains data for use case template
type UseCaseTemplateData struct {
	PackageName  string
	UseCaseName  string
	Dependencies []Dependency
	InputType    string
	OutputType   string
	ModulePath   string
	HasLogger    bool
}

// HandlerTemplateData contains data for handler template (Fiber-based)
type HandlerTemplateData struct {
	PackageName        string
	HandlerName        string
	UseCases           []Dependency
	Routes             []Route
	HasValidation      bool
	ModulePath         string
	EnableCORS         bool
	EnableRateLimiter  bool
	EnableErrorHandler bool
}

// AdapterTemplateData contains data for adapter templates (includes fx Module)
type AdapterTemplateData struct {
	PackageName   string
	InterfaceName string
	DatabaseType  string // postgres, mongodb, dynamodb
	EntityName    string
	Fields        []Field
	Methods       []Method
	ModulePath    string
}

// ConsumerTemplateData contains data for message consumer template
type ConsumerTemplateData struct {
	PackageName  string
	ConsumerName string
	UseCases     []Dependency
	ModulePath   string
	Broker       string
	Pattern      string
	Queue        string
}

// ApiClientTemplateData contains data for API client template
type ApiClientTemplateData struct {
	PackageName string
	ClientName  string
	BaseURL     string
	ModulePath  string
}

// PublisherTemplateData contains data for message publisher template
type PublisherTemplateData struct {
	PackageName   string
	PublisherName string
	ModulePath    string
	Broker        string
	Pattern       string
	Target        string
}

// MessagingFoundationTemplateData drives the per-broker infrastructure module
// that exposes the shared client (rabbit.Application or mq.Client) used by all
// generated consumers and publishers.
type MessagingFoundationTemplateData struct {
	PackageName string
	ModulePath  string
	Broker      string
}

// CacheTemplateData contains data for cache adapter template
type CacheTemplateData struct {
	PackageName string
	CacheName   string
	ModulePath  string
}

// ProjectStructure defines the directory layout
type ProjectStructure struct {
	RootDir     string
	Name        string
	ModulePath  string
	Directories []string
	Files       []ProjectFile
}

// ProjectFile represents a file to be created
type ProjectFile struct {
	Path     string
	Template string
	Data     interface{}
}

// StandardDirectories defines the standard project directory structure
var StandardDirectories = []string{
	"cmd",
	"internal/domain",
	"internal/entrypoints/http",
	"internal/entrypoints/messaging",
	"internal/drivenadapters/database",
	"internal/drivenadapters/api",
	"internal/drivenadapters/messaging",
	"internal/drivenadapters/cache",
	"config",
	"pkg",
}

// Config represents CLI tool configuration
type Config struct {
	Templates  TemplatesConfig
	Defaults   DefaultsConfig
	Formatting FormattingConfig
}

// TemplatesConfig configures template behavior
type TemplatesConfig struct {
	Directory string
	Embedded  bool
}

// DefaultsConfig provides default values
type DefaultsConfig struct {
	ModulePrefix string
	Author       string
	License      string
}

// FormattingConfig configures code formatting
type FormattingConfig struct {
	UseGofmt     bool
	UseGoimports bool
	TabWidth     int
}
