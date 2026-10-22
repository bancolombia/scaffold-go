package generator

import (
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// genValidPublisherName generates valid publisher names (PascalCase).
func genValidPublisherName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{2,20}`)
}

// genRabbitMQPublisherPattern generates one of the supported RabbitMQ patterns
// for publishers.
func genRabbitMQPublisherPattern() gopter.Gen {
	return gen.OneConstOf("event", "command", "query", "notification")
}

// TestProperty_GeneratedPublisherAdaptersAreWiredAsFxModules ensures every
// broker/pattern combination emits a per-component fx module exposing the
// `<Name>Module` variable that cmd/main.go can wire with no further edits.
func TestProperty_GeneratedPublisherAdaptersAreWiredAsFxModules(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("rabbitmq publishers always emit a fx module file", prop.ForAll(
		func(publisherName string, pattern string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				return false
			}

			g := NewPublisherGenerator(publisherName, "rabbitmq", pattern, "", tmpDir, templates)
			if err := g.Validate(); err != nil {
				return false
			}

			files, err := g.Generate()
			if err != nil {
				return false
			}

			var moduleContent string
			for _, f := range files {
				if strings.HasSuffix(f.Path, "module.go") && strings.Contains(f.Path, "internal/drivenadapters/messaging") {
					moduleContent = f.Content
					break
				}
			}
			if moduleContent == "" {
				return false
			}

			if !strings.Contains(moduleContent, "package publishers") {
				return false
			}
			if !strings.Contains(moduleContent, "go.uber.org/fx") {
				return false
			}
			if !strings.Contains(moduleContent, "var Module = fx.Options") {
				return false
			}
			if !strings.Contains(moduleContent, "fx.Provide(") {
				return false
			}
			if !strings.Contains(moduleContent, "fx.Annotate(") {
				return false
			}
			if !strings.Contains(moduleContent, "New"+publisherName+"Publisher") {
				return false
			}

			return true
		},
		genValidPublisherName(),
		genRabbitMQPublisherPattern(),
	))

	properties.TestingRun(t)
}

// TestProperty_GeneratedPublisherIncludesBrokerSpecificCalls checks that each
// generated publisher references the right broker library entry point so the
// scaffolded code is broker-aware out of the box.
func TestProperty_GeneratedPublisherIncludesBrokerSpecificCalls(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	rabbitExpected := map[string]string{
		"event":        "EventBus().Emit",
		"command":      "Gateway().SendCommand",
		"query":        "Gateway().RequestReply",
		"notification": "EventBus().EmitNotification",
	}

	properties.Property("rabbitmq publisher references the matching reactive-commons-go API", prop.ForAll(
		func(publisherName string, pattern string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				return false
			}

			g := NewPublisherGenerator(publisherName, "rabbitmq", pattern, "", tmpDir, templates)
			if err := g.Validate(); err != nil {
				return false
			}
			files, err := g.Generate()
			if err != nil {
				return false
			}

			var publisherContent string
			for _, f := range files {
				if strings.Contains(f.Path, "drivenadapters/messaging") && strings.HasSuffix(f.Path, "_publisher.go") {
					publisherContent = f.Content
					break
				}
			}
			if publisherContent == "" {
				return false
			}

			if !strings.Contains(publisherContent, "reactive-commons-go") {
				return false
			}
			expected, ok := rabbitExpected[pattern]
			if !ok {
				return false
			}
			return strings.Contains(publisherContent, expected)
		},
		genValidPublisherName(),
		genRabbitMQPublisherPattern(),
	))

	properties.TestingRun(t)
}
