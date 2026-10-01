package generator

import (
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// genValidConsumerName generates valid consumer names (PascalCase)
func genValidConsumerName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{0,30}`)
}

// genConsumerUseCaseSpec generates valid use case specifications in "name:Type" format.
func genConsumerUseCaseSpec() gopter.Gen {
	return gen.SliceOf(gen.OneGenOf(
		gen.Const("processOrder:ProcessOrderUseCase"),
		gen.Const("validatePayment:ValidatePaymentUseCase"),
		gen.Const("sendNotification:SendNotificationUseCase"),
		gen.Const("updateInventory:UpdateInventoryUseCase"),
		gen.Const("createCustomer:CreateCustomerUseCase"),
	))
}

// genRabbitMQConsumerPattern generates one of the supported RabbitMQ patterns.
func genRabbitMQConsumerPattern() gopter.Gen {
	return gen.OneConstOf("event", "command", "query", "notification")
}

// TestProperty_GeneratedConsumerEntrypointsAreWiredAsFxModules ensures every
// broker/pattern combination emits a per-component fx module exposing the
// `<Name>Module` variable that cmd/main.go can wire with no further edits.
func TestProperty_GeneratedConsumerEntrypointsAreWiredAsFxModules(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("rabbitmq consumers always emit a fx module file", prop.ForAll(
		func(consumerName string, pattern string, useCases []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewConsumerGenerator(consumerName, "rabbitmq", pattern, useCases, "", tmpDir, templates)
			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			var moduleContent string
			for _, f := range files {
				if strings.HasSuffix(f.Path, "_module.go") && strings.Contains(f.Path, "internal/entrypoints/messaging") {
					moduleContent = f.Content
					break
				}
			}
			if moduleContent == "" {
				t.Logf("per-component module file not found")
				return false
			}

			if !strings.Contains(moduleContent, "go.uber.org/fx") {
				return false
			}
			if !strings.Contains(moduleContent, "var "+consumerName+"Module") {
				return false
			}
			if !strings.Contains(moduleContent, "fx.Provide(New"+consumerName+"Consumer)") {
				return false
			}
			if !strings.Contains(moduleContent, "package consumers") {
				return false
			}
			return true
		},
		genValidConsumerName(),
		genRabbitMQConsumerPattern(),
		genConsumerUseCaseSpec(),
	))

	properties.TestingRun(t)
}

// TestProperty_GeneratedConsumerIncludesBrokerSpecificCalls checks that each
// generated consumer references the right broker library entry point so the
// scaffolded code is broker-aware out of the box.
func TestProperty_GeneratedConsumerIncludesBrokerSpecificCalls(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	rabbitExpected := map[string]string{
		"event":        "ListenEvent",
		"command":      "ListenCommand",
		"query":        "ServeQuery",
		"notification": "ListenNotification",
	}

	properties.Property("rabbitmq consumer references the matching reactive-commons-go API", prop.ForAll(
		func(consumerName string, pattern string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				return false
			}

			g := NewConsumerGenerator(consumerName, "rabbitmq", pattern, nil, "", tmpDir, templates)
			if err := g.Validate(); err != nil {
				return false
			}
			files, err := g.Generate()
			if err != nil {
				return false
			}

			var consumerContent string
			for _, f := range files {
				if strings.HasSuffix(f.Path, "_consumer.go") {
					consumerContent = f.Content
					break
				}
			}
			if consumerContent == "" {
				return false
			}

			if !strings.Contains(consumerContent, "reactive-commons-go") {
				return false
			}
			expected, ok := rabbitExpected[pattern]
			if !ok {
				return false
			}
			return strings.Contains(consumerContent, expected)
		},
		genValidConsumerName(),
		genRabbitMQConsumerPattern(),
	))

	properties.TestingRun(t)
}
