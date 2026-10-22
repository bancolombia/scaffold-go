package generator

import (
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumerGenerator_Validate(t *testing.T) {
	tests := []struct {
		name         string
		consumerName string
		broker       string
		pattern      string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid rabbitmq event consumer",
			consumerName: "OrderConsumer",
			broker:       "rabbitmq",
			pattern:      "event",
			wantErr:      false,
		},
		{
			name:         "valid ibmmq listener consumer",
			consumerName: "OrderConsumer",
			broker:       "ibmmq",
			pattern:      "listener",
			wantErr:      false,
		},
		{
			name:         "invalid consumer name with hyphen",
			consumerName: "Order-Consumer",
			broker:       "rabbitmq",
			pattern:      "event",
			wantErr:      true,
			errContains:  "invalid consumer name",
		},
		{
			name:         "missing broker",
			consumerName: "OrderConsumer",
			broker:       "",
			pattern:      "event",
			wantErr:      true,
			errContains:  "--broker is required",
		},
		{
			name:         "missing pattern",
			consumerName: "OrderConsumer",
			broker:       "rabbitmq",
			pattern:      "",
			wantErr:      true,
			errContains:  "--pattern is required",
		},
		{
			name:         "kafka rejected as not implemented",
			consumerName: "OrderConsumer",
			broker:       "kafka",
			pattern:      "event",
			wantErr:      true,
			errContains:  "kafka",
		},
		{
			name:         "unsupported broker",
			consumerName: "OrderConsumer",
			broker:       "nats",
			pattern:      "event",
			wantErr:      true,
			errContains:  "unsupported broker",
		},
		{
			name:         "rabbitmq rejects ibmmq pattern",
			consumerName: "OrderConsumer",
			broker:       "rabbitmq",
			pattern:      "listener",
			wantErr:      true,
			errContains:  "unsupported pattern",
		},
		{
			name:         "ibmmq rejects rabbitmq pattern",
			consumerName: "OrderConsumer",
			broker:       "ibmmq",
			pattern:      "event",
			wantErr:      true,
			errContains:  "unsupported pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			g := NewConsumerGenerator(tt.consumerName, tt.broker, tt.pattern, []string{}, "", "/tmp/test", templates)
			err = g.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConsumerGenerator_Generate_RabbitMQ(t *testing.T) {
	tests := []struct {
		name         string
		consumerName string
		pattern      string
		useCases     []string
		queue        string
		checkContent []string
	}{
		{
			name:         "rabbitmq event consumer",
			consumerName: "OrderEventConsumer",
			pattern:      "event",
			useCases:     []string{"createOrder:CreateOrderUseCase"},
			queue:        "order.created",
			checkContent: []string{
				"package consumers",
				"reactive-commons-go",
				"DomainEvent",
				"ListenEvent",
				"\"order.created\"",
				"createOrder CreateOrderUseCase",
				"type OrderEventConsumerConsumer struct",
				"func RegisterOrderEventConsumer",
			},
		},
		{
			name:         "rabbitmq command consumer",
			consumerName: "InvoiceCommandConsumer",
			pattern:      "command",
			useCases:     []string{},
			queue:        "send-invoice",
			checkContent: []string{
				"async.Command",
				"ListenCommand",
				"\"send-invoice\"",
			},
		},
		{
			name:         "rabbitmq query consumer",
			consumerName: "ProductQueryConsumer",
			pattern:      "query",
			useCases:     []string{},
			queue:        "get-product",
			checkContent: []string{
				"AsyncQuery",
				"ServeQuery",
				"async.From",
			},
		},
		{
			name:         "rabbitmq notification consumer",
			consumerName: "CacheConsumer",
			pattern:      "notification",
			useCases:     []string{},
			queue:        "cache-invalidated",
			checkContent: []string{
				"Notification",
				"ListenNotification",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			g := NewConsumerGenerator(tt.consumerName, "rabbitmq", tt.pattern, tt.useCases, tt.queue, "/tmp/test", templates)
			files, err := g.Generate()
			require.NoError(t, err)

			// Should produce at least the consumer + per-component module + foundation module.
			assert.GreaterOrEqual(t, len(files), 3, "should emit consumer + module + foundation")

			var consumerContent string
			for _, f := range files {
				if strings.HasSuffix(f.Path, "_consumer.go") {
					consumerContent = f.Content
					assert.Contains(t, f.Path, "internal/entrypoints/messaging")
				}
			}

			require.NotEmpty(t, consumerContent)
			for _, want := range tt.checkContent {
				assert.Contains(t, consumerContent, want)
			}
		})
	}
}

func TestConsumerGenerator_Generate_IBMMQ(t *testing.T) {
	tests := []struct {
		name         string
		consumerName string
		pattern      string
		queue        string
		checkContent []string
	}{
		{
			name:         "ibmmq listener",
			consumerName: "InvoiceListener",
			pattern:      "listener",
			queue:        "DEV.QUEUE.INVOICES",
			checkContent: []string{
				"package consumers",
				"commons-mq-go/mq",
				"NewListener",
				"\"DEV.QUEUE.INVOICES\"",
			},
		},
		{
			name:         "ibmmq requestreply",
			consumerName: "QueryServer",
			pattern:      "requestreply",
			queue:        "DEV.QUEUE.QUERY",
			checkContent: []string{
				"commons-mq-go/mq",
				"NewListener",
				"ReplyToQueue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			g := NewConsumerGenerator(tt.consumerName, "ibmmq", tt.pattern, []string{}, tt.queue, "/tmp/test", templates)
			files, err := g.Generate()
			require.NoError(t, err)

			var consumerContent string
			for _, f := range files {
				if strings.HasSuffix(f.Path, "_consumer.go") {
					consumerContent = f.Content
				}
			}

			require.NotEmpty(t, consumerContent)
			for _, want := range tt.checkContent {
				assert.Contains(t, consumerContent, want)
			}
		})
	}
}

func TestConsumerGenerator_PerComponentModule(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	g := NewConsumerGenerator("OrderConsumer", "rabbitmq", "event", nil, "order.created", "/tmp/test", templates)
	files, err := g.Generate()
	require.NoError(t, err)

	// One file should be the per-component module, named after the component
	// (NOT a shared "module.go" that would be overwritten on the next generate).
	// Match on the entrypoints path specifically: the foundation module under
	// internal/infrastructure/messaging also ends with "module.go".
	var moduleContent, modulePath string
	for _, f := range files {
		if strings.Contains(f.Path, "entrypoints/messaging/rabbitmq") && strings.HasSuffix(f.Path, "module.go") {
			moduleContent = f.Content
			modulePath = f.Path
		}
	}
	require.NotEmpty(t, moduleContent)
	assert.Contains(t, modulePath, "module.go")
	assert.Contains(t, moduleContent, "package consumers")
	assert.Contains(t, moduleContent, "var Module")
	assert.Contains(t, moduleContent, "fx.Provide(NewOrderConsumerConsumer)")
	assert.Contains(t, moduleContent, "fx.Invoke(RegisterOrderConsumer)")
}

func TestConsumerGenerator_FoundationModule(t *testing.T) {
	tmpDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	g := NewConsumerGenerator("OrderConsumer", "rabbitmq", "event", nil, "order.created", tmpDir, templates)
	files, err := g.Generate()
	require.NoError(t, err)

	var foundationContent string
	for _, f := range files {
		if strings.Contains(f.Path, "infrastructure/messaging/rabbitmq") {
			foundationContent = f.Content
		}
	}
	require.NotEmpty(t, foundationContent)
	assert.Contains(t, foundationContent, "package rabbitmqinfra")
	assert.Contains(t, foundationContent, "rabbit.NewApplication")
	assert.Contains(t, foundationContent, "fx.Module(\"rabbitmq-infra\"")
}

func TestConsumerGenerator_ComponentType(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	g := NewConsumerGenerator("OrderConsumer", "rabbitmq", "event", nil, "order.created", "/tmp/test", templates)
	assert.Equal(t, "consumer", g.ComponentType())
}

func TestConsumerGenerator_ParseUseCases(t *testing.T) {
	tests := []struct {
		name     string
		useCases []string
		want     int
		checkDep func(t *testing.T, deps []template.Dependency)
	}{
		{
			name:     "empty use cases",
			useCases: []string{},
			want:     0,
		},
		{
			name:     "single use case",
			useCases: []string{"createOrder:CreateOrderUseCase"},
			want:     1,
			checkDep: func(t *testing.T, deps []template.Dependency) {
				assert.Equal(t, "createOrder", deps[0].Name)
				assert.Equal(t, "CreateOrderUseCase", deps[0].Type)
			},
		},
		{
			name:     "use case with package alias",
			useCases: []string{"createOrder:domain.CreateOrderUseCase:domain"},
			want:     1,
			checkDep: func(t *testing.T, deps []template.Dependency) {
				assert.Equal(t, "createOrder", deps[0].Name)
				assert.Equal(t, "domain.CreateOrderUseCase", deps[0].Type)
				assert.Equal(t, "domain", deps[0].Package)
			},
		},
		{
			name:     "invalid format - no colon",
			useCases: []string{"createOrder"},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			g := NewConsumerGenerator("OrderConsumer", "rabbitmq", "event", tt.useCases, "order", "/tmp/test", templates)
			deps := g.parseUseCases()

			assert.Len(t, deps, tt.want)
			if tt.checkDep != nil {
				tt.checkDep(t, deps)
			}
		})
	}
}
