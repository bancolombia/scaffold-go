package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gabheadz/scaffold-go/internal/template"
	"github.com/gabheadz/scaffold-go/internal/validator"
)

// Supported messaging brokers for ep consumer / da publisher.
const (
	BrokerRabbitMQ = "rabbitmq"
	BrokerIBMMQ    = "ibmmq"
	BrokerKafka    = "kafka"
)

// ConsumerPackageName is the Go package name used for files emitted under
// internal/entrypoints/messaging. Renamed from "messaging" to avoid the alias
// collision with the publisher half when both are imported in cmd/main.go.
const ConsumerPackageName = "consumers"

// PublisherPackageName is the Go package name used for files emitted under
// internal/drivenadapters/messaging.
const PublisherPackageName = "publishers"

// supportedConsumerPatterns enumerates valid (broker, pattern) combinations
// for the entry-point consumer half.
var supportedConsumerPatterns = map[string]map[string]struct{}{
	BrokerRabbitMQ: {
		"event":        {},
		"command":      {},
		"query":        {},
		"notification": {},
	},
	BrokerIBMMQ: {
		"listener":     {},
		"requestreply": {},
	},
}

// ConsumerGenerator emits a broker-specific message consumer plus its per-component
// fx module, and (idempotently) the foundation module that provides the shared
// broker client to all consumers/publishers.
type ConsumerGenerator struct {
	Name       string
	Broker     string
	Pattern    string
	UseCases   []string
	Queue      string
	ProjectDir string
	Templates  *template.TemplateEngine
	Validator  *validator.Validator
}

// NewConsumerGenerator constructs a broker-aware consumer generator.
func NewConsumerGenerator(name, broker, pattern string, useCases []string, queue, projectDir string, templates *template.TemplateEngine) *ConsumerGenerator {
	return &ConsumerGenerator{
		Name:       name,
		Broker:     strings.ToLower(strings.TrimSpace(broker)),
		Pattern:    strings.ToLower(strings.TrimSpace(pattern)),
		UseCases:   useCases,
		Queue:      strings.TrimSpace(queue),
		ProjectDir: projectDir,
		Templates:  templates,
		Validator:  validator.New(),
	}
}

// Validate enforces broker/pattern combos and rejects Kafka with a helpful error.
func (g *ConsumerGenerator) Validate() error {
	if err := g.Validator.ValidateName(g.Name); err != nil {
		return fmt.Errorf("invalid consumer name: %w", err)
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
	patterns, ok := supportedConsumerPatterns[g.Broker]
	if !ok {
		return fmt.Errorf("unsupported broker %q (supported: rabbitmq, ibmmq, kafka)", g.Broker)
	}
	if _, ok := patterns[g.Pattern]; !ok {
		return fmt.Errorf("unsupported pattern %q for broker %q (supported: %s)", g.Pattern, g.Broker, joinKeys(patterns))
	}
	return nil
}

// Generate renders the consumer file, the per-component fx module, and (if
// missing) the shared broker foundation module.
func (g *ConsumerGenerator) Generate() ([]GeneratedFile, error) {
	queue := g.Queue
	if queue == "" {
		queue = strings.ToLower(g.Name)
	}

	modulePath, err := g.Validator.GetModulePath(g.ProjectDir)
	if err != nil {
		modulePath = ""
	}

	data := template.ConsumerTemplateData{
		PackageName:  ConsumerPackageName,
		ConsumerName: g.Name,
		UseCases:     g.parseUseCases(),
		ModulePath:   modulePath,
		Broker:       g.Broker,
		Pattern:      g.Pattern,
		Queue:        queue,
	}

	consumerTemplate := fmt.Sprintf("ep/consumer/%s/%s_consumer.go.tmpl", g.Broker, g.Pattern)
	consumerContent, err := g.Templates.Execute(consumerTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate consumer: %w", err)
	}

	moduleContent, err := g.Templates.Execute("ep/consumer/component_module.go.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate component module: %w", err)
	}

	consumerPath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "messaging", strings.ToLower(g.Name)+"_consumer.go")
	moduleFilePath := filepath.Join(g.ProjectDir, "internal", "entrypoints", "messaging", strings.ToLower(g.Name)+"_module.go")

	files := []GeneratedFile{
		{Path: consumerPath, Content: consumerContent},
		{Path: moduleFilePath, Content: moduleContent},
	}

	if foundation, ok, err := renderFoundationModule(g.Templates, g.Broker, g.ProjectDir, modulePath); err != nil {
		return nil, err
	} else if ok {
		files = append(files, foundation)
	}

	return files, nil
}

// ComponentType reports the kind to error/log helpers.
func (g *ConsumerGenerator) ComponentType() string {
	return "consumer"
}

// parseUseCases turns the `name:Type[:Package]` slice into structured
// dependencies, populating the optional Package field used by the consumer
// templates so that imports can be qualified correctly.
func (g *ConsumerGenerator) parseUseCases() []template.Dependency {
	out := make([]template.Dependency, 0, len(g.UseCases))
	for _, uc := range g.UseCases {
		parts := strings.Split(uc, ":")
		if len(parts) < 2 {
			continue
		}
		dep := template.Dependency{
			Name: strings.TrimSpace(parts[0]),
			Type: strings.TrimSpace(parts[1]),
		}
		if len(parts) >= 3 {
			dep.Package = strings.TrimSpace(parts[2])
		}
		out = append(out, dep)
	}
	return out
}

func joinKeys(m map[string]struct{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// sort to make error messages stable
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
	return strings.Join(keys, ", ")
}
