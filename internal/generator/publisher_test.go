package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublisherGenerator_Validate(t *testing.T) {
	tests := []struct {
		name          string
		publisherName string
		broker        string
		pattern       string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "valid rabbitmq event publisher",
			publisherName: "OrderEventPublisher",
			broker:        "rabbitmq",
			pattern:       "event",
			wantErr:       false,
		},
		{
			name:          "valid ibmmq sender",
			publisherName: "InvoiceSender",
			broker:        "ibmmq",
			pattern:       "sender",
			wantErr:       false,
		},
		{
			name:          "valid ibmmq requestreply",
			publisherName: "QueryClient",
			broker:        "ibmmq",
			pattern:       "requestreply",
			wantErr:       false,
		},
		{
			name:          "invalid publisher name with hyphen",
			publisherName: "Order-Event",
			broker:        "rabbitmq",
			pattern:       "event",
			wantErr:       true,
			errContains:   "invalid publisher name",
		},
		{
			name:          "missing broker",
			publisherName: "OrderEventPublisher",
			broker:        "",
			pattern:       "event",
			wantErr:       true,
			errContains:   "--broker is required",
		},
		{
			name:          "missing pattern",
			publisherName: "OrderEventPublisher",
			broker:        "rabbitmq",
			pattern:       "",
			wantErr:       true,
			errContains:   "--pattern is required",
		},
		{
			name:          "kafka rejected",
			publisherName: "OrderEventPublisher",
			broker:        "kafka",
			pattern:       "event",
			wantErr:       true,
			errContains:   "kafka",
		},
		{
			name:          "unsupported broker",
			publisherName: "OrderEventPublisher",
			broker:        "nats",
			pattern:       "event",
			wantErr:       true,
			errContains:   "unsupported broker",
		},
		{
			name:          "rabbitmq rejects ibmmq pattern",
			publisherName: "OrderEventPublisher",
			broker:        "rabbitmq",
			pattern:       "sender",
			wantErr:       true,
			errContains:   "unsupported pattern",
		},
		{
			name:          "ibmmq rejects rabbitmq pattern",
			publisherName: "OrderEventPublisher",
			broker:        "ibmmq",
			pattern:       "event",
			wantErr:       true,
			errContains:   "unsupported pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			g := NewPublisherGenerator(tt.publisherName, tt.broker, tt.pattern, "", t.TempDir(), templates)
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

func TestPublisherGenerator_Generate_RabbitMQ(t *testing.T) {
	tests := []struct {
		name          string
		publisherName string
		pattern       string
		target        string
		checkContent  []string
	}{
		{
			name:          "event",
			publisherName: "OrderEventPublisher",
			pattern:       "event",
			target:        "order.created",
			checkContent: []string{
				"package publishers",
				"reactive-commons-go",
				"async.DomainEvent",
				"EventBus().Emit(",
				"\"order.created\"",
				"type OrderEventPublisherPublisher struct",
			},
		},
		{
			name:          "command",
			publisherName: "InvoiceCommand",
			pattern:       "command",
			target:        "send-invoice",
			checkContent: []string{
				"async.Command",
				"Gateway().SendCommand",
				"\"send-invoice\"",
			},
		},
		{
			name:          "query",
			publisherName: "ProductQuery",
			pattern:       "query",
			target:        "get-product",
			checkContent: []string{
				"AsyncQuery",
				"Gateway().RequestReply",
			},
		},
		{
			name:          "notification",
			publisherName: "CacheNotify",
			pattern:       "notification",
			target:        "cache-invalidated",
			checkContent: []string{
				"async.Notification",
				"EventBus().EmitNotification",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			g := NewPublisherGenerator(tt.publisherName, "rabbitmq", tt.pattern, tt.target, t.TempDir(), templates)
			files, err := g.Generate()
			require.NoError(t, err)

			assert.GreaterOrEqual(t, len(files), 3, "should emit publisher + module + foundation")

			var publisherContent string
			for _, f := range files {
				if strings.Contains(f.Path, "drivenadapters/messaging") && strings.HasSuffix(f.Path, "_publisher.go") {
					publisherContent = f.Content
					assert.Contains(t, f.Path, "internal/drivenadapters/messaging")
				}
			}

			require.NotEmpty(t, publisherContent)
			for _, want := range tt.checkContent {
				assert.Contains(t, publisherContent, want)
			}
		})
	}
}

func TestPublisherGenerator_Generate_IBMMQ(t *testing.T) {
	tests := []struct {
		name          string
		publisherName string
		pattern       string
		target        string
		checkContent  []string
	}{
		{
			name:          "sender",
			publisherName: "InvoiceSender",
			pattern:       "sender",
			target:        "DEV.QUEUE.INVOICES",
			checkContent: []string{
				"package publishers",
				"commons-mq-go/mq",
				"NewSender",
				"\"DEV.QUEUE.INVOICES\"",
			},
		},
		{
			name:          "requestreply",
			publisherName: "QueryClient",
			pattern:       "requestreply",
			target:        "DEV.QUEUE.QUERY",
			checkContent: []string{
				"NewRequestReplyClient",
				"\"DEV.QUEUE.QUERY\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			g := NewPublisherGenerator(tt.publisherName, "ibmmq", tt.pattern, tt.target, t.TempDir(), templates)
			files, err := g.Generate()
			require.NoError(t, err)

			var publisherContent string
			for _, f := range files {
				if strings.HasSuffix(f.Path, "_publisher.go") {
					publisherContent = f.Content
				}
			}

			require.NotEmpty(t, publisherContent)
			for _, want := range tt.checkContent {
				assert.Contains(t, publisherContent, want)
			}
		})
	}
}

func TestPublisherGenerator_PerComponentModule(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	g := NewPublisherGenerator("OrderEventPublisher", "rabbitmq", "event", "order.created", t.TempDir(), templates)
	files, err := g.Generate()
	require.NoError(t, err)

	var moduleContent, modulePath string
	for _, f := range files {
		if strings.HasSuffix(f.Path, "module.go") && strings.Contains(f.Path, "drivenadapters/messaging/rabbitmq/") {
			moduleContent = f.Content
			modulePath = f.Path
		}
	}
	require.NotEmpty(t, moduleContent)
	assert.Contains(t, modulePath, "module.go")
	assert.Contains(t, moduleContent, "package publishers")
	assert.Contains(t, moduleContent, "var Module")
	assert.Contains(t, moduleContent, "fx.Provide(")
	assert.Contains(t, moduleContent, "fx.Annotate(")
	assert.Contains(t, moduleContent, "NewOrderEventPublisherPublisher")
}

func TestPublisherGenerator_FoundationModuleEmittedOnce(t *testing.T) {
	tmpDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	g := NewPublisherGenerator("OrderEventPublisher", "rabbitmq", "event", "order.created", tmpDir, templates)
	files, err := g.Generate()
	require.NoError(t, err)

	// Write the foundation file to disk so the next call sees it as existing.
	foundationDir := filepath.Join(tmpDir, "internal", "infrastructure", "messaging", "rabbitmq")
	require.NoError(t, os.MkdirAll(foundationDir, 0o755))
	for _, f := range files {
		if strings.Contains(f.Path, "infrastructure/messaging/rabbitmq") {
			require.NoError(t, os.WriteFile(f.Path, []byte(f.Content), 0o644))
		}
	}

	// Second invocation should skip the foundation module entirely.
	g2 := NewPublisherGenerator("AnotherPublisher", "rabbitmq", "event", "another.created", tmpDir, templates)
	files2, err := g2.Generate()
	require.NoError(t, err)

	for _, f := range files2 {
		if strings.Contains(f.Path, "infrastructure/messaging/rabbitmq") {
			t.Fatalf("foundation module should not be re-emitted; got %s", f.Path)
		}
	}
}

func TestPublisherGenerator_ComponentType(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	g := NewPublisherGenerator("OrderEventPublisher", "rabbitmq", "event", "order.created", t.TempDir(), templates)
	assert.Equal(t, "publisher", g.ComponentType())
}
