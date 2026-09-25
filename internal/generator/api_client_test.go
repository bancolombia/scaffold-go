package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabheadz/scaffold-go/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiClientGenerator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		clientName string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid client name",
			clientName: "PaymentService",
			wantErr:    false,
		},
		{
			name:       "invalid client name with hyphen",
			clientName: "Payment-Service",
			wantErr:    true,
			errMsg:     "invalid API client name",
		},
		{
			name:       "invalid client name with space",
			clientName: "Payment Service",
			wantErr:    true,
			errMsg:     "invalid API client name",
		},
		{
			name:       "invalid client name starting with number",
			clientName: "123Payment",
			wantErr:    true,
			errMsg:     "invalid API client name",
		},
		{
			name:       "empty client name",
			clientName: "",
			wantErr:    true,
			errMsg:     "invalid API client name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			gen := NewApiClientGenerator(tt.clientName, "https://api.example.com", tempDir, templates)
			err = gen.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApiClientGenerator_Generate(t *testing.T) {
	tempDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewApiClientGenerator("PaymentService", "https://api.example.com", tempDir, templates)
	files, err := gen.Generate()

	require.NoError(t, err)
	assert.Len(t, files, 2, "should generate 2 files: client and module")

	// Verify client file
	clientFile := files[0]
	assert.Contains(t, clientFile.Path, "paymentservice_client.go")
	assert.Contains(t, clientFile.Content, "package api")
	assert.Contains(t, clientFile.Content, "type PaymentServiceClient struct")
	assert.Contains(t, clientFile.Content, "func NewPaymentServiceClient")
	assert.Contains(t, clientFile.Content, "httpClient *http.Client")
	assert.Contains(t, clientFile.Content, "logger *zap.Logger")

	// Verify HTTP methods
	assert.Contains(t, clientFile.Content, "func (c *PaymentServiceClient) Get(")
	assert.Contains(t, clientFile.Content, "func (c *PaymentServiceClient) Post(")
	assert.Contains(t, clientFile.Content, "func (c *PaymentServiceClient) Put(")
	assert.Contains(t, clientFile.Content, "func (c *PaymentServiceClient) Delete(")

	// Verify error handling
	assert.Contains(t, clientFile.Content, "if err != nil")
	assert.Contains(t, clientFile.Content, "c.logger.Error")

	// Verify retry logic
	assert.Contains(t, clientFile.Content, "maxRetries")
	assert.Contains(t, clientFile.Content, "Retrying request")

	// Verify module file
	moduleFile := files[1]
	assert.Contains(t, moduleFile.Path, "module.go")
	assert.Contains(t, moduleFile.Content, "package api")
	assert.Contains(t, moduleFile.Content, "fx.Module")
	assert.Contains(t, moduleFile.Content, "fx.Provide")
	assert.Contains(t, moduleFile.Content, "NewPaymentServiceClient")
}

func TestApiClientGenerator_ComponentType(t *testing.T) {
	tempDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewApiClientGenerator("PaymentService", "https://api.example.com", tempDir, templates)
	assert.Equal(t, "api-client", gen.ComponentType())
}

func TestApiClientGenerator_GeneratedFileStructure(t *testing.T) {
	tempDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewApiClientGenerator("UserService", "https://api.example.com", tempDir, templates)
	files, err := gen.Generate()
	require.NoError(t, err)

	// Write files to disk
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(file.Path, []byte(file.Content), 0644)
		require.NoError(t, err)
	}

	// Verify file paths
	clientPath := filepath.Join(tempDir, "internal", "drivenadapters", "api", "userservice_client.go")
	modulePath := filepath.Join(tempDir, "internal", "drivenadapters", "api", "module.go")

	assert.FileExists(t, clientPath)
	assert.FileExists(t, modulePath)

	// Verify client file content
	clientContent, err := os.ReadFile(clientPath)
	require.NoError(t, err)
	assert.Contains(t, string(clientContent), "type UserServiceClient struct")
	assert.Contains(t, string(clientContent), "func NewUserServiceClient")

	// Verify module file content
	moduleContent, err := os.ReadFile(modulePath)
	require.NoError(t, err)
	assert.Contains(t, string(moduleContent), "var Module = fx.Module")
}

func TestApiClientGenerator_HTTPMethodsIncluded(t *testing.T) {
	tempDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewApiClientGenerator("OrderService", "https://api.example.com", tempDir, templates)
	files, err := gen.Generate()
	require.NoError(t, err)

	clientFile := files[0]
	content := clientFile.Content

	// Verify all HTTP methods are present
	httpMethods := []string{"Get", "Post", "Put", "Delete"}
	for _, method := range httpMethods {
		methodSignature := "func (c *OrderServiceClient) " + method + "("
		assert.Contains(t, content, methodSignature, "should include %s method", method)
	}

	// Verify makeRequest helper
	assert.Contains(t, content, "func (c *OrderServiceClient) makeRequest(")
	assert.Contains(t, content, "http.NewRequestWithContext")

	// Verify response handling
	assert.Contains(t, content, "func (c *OrderServiceClient) decodeResponse(")
	assert.Contains(t, content, "json.NewDecoder")
}

func TestApiClientGenerator_LoggerIntegration(t *testing.T) {
	tempDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewApiClientGenerator("NotificationService", "https://api.example.com", tempDir, templates)
	files, err := gen.Generate()
	require.NoError(t, err)

	clientFile := files[0]
	content := clientFile.Content

	// Verify logger is in struct
	assert.Contains(t, content, "logger *zap.Logger")

	// Verify logger is in constructor
	assert.Contains(t, content, "func NewNotificationServiceClient(baseURL string, logger *zap.Logger)")

	// Verify logger is used for logging
	assert.Contains(t, content, "c.logger.Info")
	assert.Contains(t, content, "c.logger.Error")
	assert.Contains(t, content, "zap.String")
	assert.Contains(t, content, "zap.Error")
}

func TestApiClientGenerator_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewApiClientGenerator("InventoryService", "https://api.example.com", tempDir, templates)
	files, err := gen.Generate()
	require.NoError(t, err)

	clientFile := files[0]
	content := clientFile.Content

	// Verify error handling patterns
	assert.Contains(t, content, "if err != nil")
	assert.Contains(t, content, "return nil, err")
	assert.Contains(t, content, "return err")

	// Verify HTTP error handling
	assert.Contains(t, content, "resp.StatusCode >= 400")
	assert.Contains(t, content, "resp.StatusCode >= 500")
	assert.Contains(t, content, "fmt.Errorf")
}

func TestApiClientGenerator_RetryLogic(t *testing.T) {
	tempDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewApiClientGenerator("ShippingService", "https://api.example.com", tempDir, templates)
	files, err := gen.Generate()
	require.NoError(t, err)

	clientFile := files[0]
	content := clientFile.Content

	// Verify retry logic is present
	assert.Contains(t, content, "maxRetries")
	assert.Contains(t, content, "for attempt := 0; attempt < maxRetries; attempt++")
	assert.Contains(t, content, "Retrying request")
	assert.Contains(t, content, "time.Sleep")

	// Verify retry on server errors
	assert.Contains(t, content, "Server error, will retry")
	assert.Contains(t, content, "continue")
}

func TestApiClientGenerator_FxModuleGeneration(t *testing.T) {
	tempDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewApiClientGenerator("CatalogService", "https://api.example.com", tempDir, templates)
	files, err := gen.Generate()
	require.NoError(t, err)

	require.Len(t, files, 2)
	moduleFile := files[1]

	// Verify fx Module structure
	assert.Contains(t, moduleFile.Content, "package api")
	assert.Contains(t, moduleFile.Content, `import "go.uber.org/fx"`)
	assert.Contains(t, moduleFile.Content, "var Module = fx.Module")
	assert.Contains(t, moduleFile.Content, `fx.Module("api"`)
	assert.Contains(t, moduleFile.Content, "fx.Provide")
	assert.Contains(t, moduleFile.Content, "NewCatalogServiceClient")
}

func TestApiClientGenerator_MultipleClients(t *testing.T) {
	tempDir := t.TempDir()
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	clients := []string{"AuthService", "BillingService", "AnalyticsService"}

	for _, clientName := range clients {
		gen := NewApiClientGenerator(clientName, "https://api.example.com", tempDir, templates)
		files, err := gen.Generate()
		require.NoError(t, err)

		clientFile := files[0]

		// Verify each client has unique name
		assert.Contains(t, clientFile.Content, "type "+clientName+"Client struct")
		assert.Contains(t, clientFile.Content, "func New"+clientName+"Client")

		// Verify file naming
		expectedFileName := strings.ToLower(clientName) + "_client.go"
		assert.Contains(t, clientFile.Path, expectedFileName)
	}
}
