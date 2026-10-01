package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/bancolombia/scaffold-go/internal/validator"
)

// supportedPublisherPatterns enumerates valid (broker, pattern) combinations
// for the driven-adapter publisher half.
var supportedPublisherPatterns = map[string]map[string]struct{}{
	BrokerRabbitMQ: {
		"event":        {},
		"command":      {},
		"query":        {},
		"notification": {},
	},
	BrokerIBMMQ: {
		"sender":       {},
		"requestreply": {},
	},
}

// PublisherGenerator emits a broker-specific message publisher plus its
// per-component fx module, and (idempotently) the foundation module that
// provides the shared broker client to all consumers/publishers.
type PublisherGenerator struct {
	Name       string
	Broker     string
	Pattern    string
	Target     string
	ProjectDir string
	Templates  *template.TemplateEngine
	Validator  *validator.Validator
}

// NewPublisherGenerator constructs a broker-aware publisher generator.
func NewPublisherGenerator(name, broker, pattern, target, projectDir string, templates *template.TemplateEngine) *PublisherGenerator {
	return &PublisherGenerator{
		Name:       name,
		Broker:     strings.ToLower(strings.TrimSpace(broker)),
		Pattern:    strings.ToLower(strings.TrimSpace(pattern)),
		Target:     strings.TrimSpace(target),
		ProjectDir: projectDir,
		Templates:  templates,
		Validator:  validator.New(),
	}
}

// Validate enforces broker/pattern combos and rejects Kafka with a helpful error.
func (g *PublisherGenerator) Validate() error {
	if err := g.Validator.ValidateName(g.Name); err != nil {
		return fmt.Errorf("invalid publisher name: %w", err)
	}
	if g.Broker == "" {
		return fmt.Errorf("--broker is required (rabbitmq, ibmmq, or kafka)")
	}
	if g.Pattern == "" {
		return fmt.Errorf("--pattern is required for broker %s", g.Broker)
	}
	if g.Broker == BrokerKafka {
		return fmt.Errorf("kafka broker is reserved as a placeholder and not implemented yet")
	}
	patterns, ok := supportedPublisherPatterns[g.Broker]
	if !ok {
		return fmt.Errorf("unsupported broker %q (supported: rabbitmq, ibmmq, kafka)", g.Broker)
	}
	if _, ok := patterns[g.Pattern]; !ok {
		return fmt.Errorf("unsupported pattern %q for broker %q (supported: %s)", g.Pattern, g.Broker, joinKeys(patterns))
	}
	return nil
}

// Generate renders the publisher file, the per-component fx module, and (if
// missing) the shared broker foundation module.
func (g *PublisherGenerator) Generate() ([]GeneratedFile, error) {
	target := g.Target
	if target == "" {
		target = strings.ToLower(g.Name)
	}

	modulePath, err := g.Validator.GetModulePath(g.ProjectDir)
	if err != nil {
		modulePath = ""
	}

	data := template.PublisherTemplateData{
		PackageName:   PublisherPackageName,
		PublisherName: g.Name,
		ModulePath:    modulePath,
		Broker:        g.Broker,
		Pattern:       g.Pattern,
		Target:        target,
	}

	// Initialize the files slice to collect generated files
	var files []GeneratedFile

	// 1. Define the interface in the domain layer for the publisher.
	domainPublisherTemplate := "domain/entity/publisher.go.tmpl"
	domainPublisherContent, err := g.Templates.Execute(domainPublisherTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate domain publisher interface: %w", err)
	}
	domainPublisherPath := filepath.Join(g.ProjectDir, "internal", "domain", strings.ToLower(g.Pattern)+"_publisher.go")
	files = append(files, GeneratedFile{Path: domainPublisherPath, Content: domainPublisherContent})

	// 2. Create the foundation module if it doesn't exist as an infrastructure module to serve both publishers and consumers.
	if foundation, ok, err := renderFoundationModule(g.Templates, g.Broker, g.ProjectDir, modulePath); err != nil {
		return nil, err
	} else if ok {
		files = append(files, foundation)
	}

	// 3. Generate the publisher implementation based on the broker and pattern.
	publisherTemplate := fmt.Sprintf("da/publisher/%s/%s_publisher.go.tmpl", g.Broker, g.Pattern)
	publisherContent, err := g.Templates.Execute(publisherTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate publisher: %w", err)
	}

	// Generate the component module content
	moduleContent, err := g.Templates.Execute("da/publisher/component_module.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate component module: %w", err)
	}

	publisherPath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "messaging", g.Broker, strings.ToLower(g.Name)+"_publisher.go")
	moduleFilePath := filepath.Join(g.ProjectDir, "internal", "drivenadapters", "messaging", g.Broker, "module.go")

	files = append(files,
		GeneratedFile{Path: publisherPath, Content: publisherContent},
		GeneratedFile{Path: moduleFilePath, Content: moduleContent},
	)

	return files, nil
}

// ComponentType reports the kind to error/log helpers.
func (g *PublisherGenerator) ComponentType() string {
	return "publisher"
}

// renderFoundationModule lazily emits internal/infrastructure/messaging/<broker>/module.go.
// Returns ok=false if the file already exists on disk so callers don't add a
// duplicate (idempotent on regeneration).
func renderFoundationModule(tmpl *template.TemplateEngine, broker, projectDir, modulePath string) (GeneratedFile, bool, error) {
	dir := filepath.Join(projectDir, "internal", "infrastructure", "messaging", broker)
	path := filepath.Join(dir, "module.go")
	if exists(path) {
		return GeneratedFile{}, false, nil
	}

	pkg := foundationPackageName(broker)
	data := template.MessagingFoundationTemplateData{
		PackageName: pkg,
		ModulePath:  modulePath,
		Broker:      broker,
	}

	content, err := tmpl.Execute(fmt.Sprintf("infrastructure/messaging/%s/module.go.tmpl", broker), data)
	if err != nil {
		return GeneratedFile{}, false, fmt.Errorf("failed to generate foundation module: %w", err)
	}

	return GeneratedFile{Path: path, Content: content}, true, nil
}

// foundationPackageName returns the Go package name used by the broker-specific
// infrastructure module. Aliasing the brokers with `infra` keeps the imports in
// cmd/main.go unambiguous.
func foundationPackageName(broker string) string {
	switch broker {
	case BrokerRabbitMQ:
		return "rabbitmqinfra"
	case BrokerIBMMQ:
		return "ibmmqinfra"
	}
	return broker + "infra"
}

// FoundationInfo describes the importable broker foundation module that
// commands need when wiring cmd/main.go.
type FoundationInfo struct {
	// Alias is the Go package identifier exported by the foundation module
	// (e.g. "rabbitmqinfra"). Empty if the broker has no foundation module
	// (e.g. Kafka placeholder).
	Alias string
}

// FoundationImportInfo returns the import alias for the broker's
// internal/infrastructure/messaging/<broker> package, or a zero value if the
// broker has no foundation module.
func FoundationImportInfo(broker string) FoundationInfo {
	switch broker {
	case BrokerRabbitMQ, BrokerIBMMQ:
		return FoundationInfo{Alias: foundationPackageName(broker)}
	}
	return FoundationInfo{}
}
