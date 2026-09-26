package template

import (
	"strings"
	"testing"
)

func TestNewTemplateEngine(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine() failed: %v", err)
	}
	if engine == nil {
		t.Fatal("NewTemplateEngine() returned nil engine")
	}
	if engine.templates == nil {
		t.Fatal("NewTemplateEngine() returned engine with nil templates")
	}
}

func TestTemplateEngine_Execute(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine() failed: %v", err)
	}

	tests := []struct {
		name         string
		templateName string
		data         interface{}
		wantErr      bool
		wantContains []string
	}{
		{
			name:         "entity template",
			templateName: "domain/entity/entity.go.tmpl",
			data: EntityTemplateData{
				PackageName: "domain",
				EntityName:  "Customer",
				Fields: []Field{
					{Name: "ID", Type: "string"},
					{Name: "Name", Type: "string"},
				},
				ModulePath: "github.com/test/project",
			},
			wantErr: false,
			wantContains: []string{
				"type Customer struct",
				"ID string",
				"Name string",
			},
		},
		{
			name:         "repository template",
			templateName: "domain/entity/repository.go.tmpl",
			data: EntityTemplateData{
				PackageName:    "domain",
				EntityName:     "Customer",
				RepositoryName: "CustomerRepository",
				ModulePath:     "github.com/test/project",
			},
			wantErr: false,
			wantContains: []string{
				"type CustomerRepository interface",
				"Create",
				"GetByID",
				"Update",
				"Delete",
				"List",
			},
		},
		{
			name:         "usecase template",
			templateName: "domain/usecase/usecase.go.tmpl",
			data: UseCaseTemplateData{
				PackageName: "domain",
				UseCaseName: "CreateCustomer",
				Dependencies: []Dependency{
					{Name: "customerRepo", Type: "CustomerRepository"},
				},
				ModulePath: "github.com/test/project",
			},
			wantErr: false,
			wantContains: []string{
				"type CreateCustomerUseCase interface",
				"Execute",
				"customerRepo domain.CustomerRepository",
			},
		},
		{
			name:         "template not found",
			templateName: "missing/nonexistent.tmpl",
			data:         nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Execute(tt.templateName, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, want := range tt.wantContains {
					if !strings.Contains(result, want) {
						t.Errorf("Execute() result does not contain %q\nGot:\n%s", want, result)
					}
				}
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PascalCase", "pascal_case"},
		{"camelCase", "camel_case"},
		{"HTTPServer", "h_t_t_p_server"},
		{"Customer", "customer"},
		{"CustomerID", "customer_i_d"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PascalCase", "pascalCase"},
		{"Customer", "customer"},
		{"HTTPServer", "hTTPServer"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toCamelCase(tt.input)
			if got != tt.want {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PascalCase", "pascal-case"},
		{"camelCase", "camel-case"},
		{"Customer", "customer"},
		{"CustomerID", "customer-i-d"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toKebabCase(tt.input)
			if got != tt.want {
				t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"customer", "customers"},
		{"entity", "entities"},
		{"class", "classes"},
		{"person", "persons"},
		{"", "s"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := pluralize(tt.input)
			if got != tt.want {
				t.Errorf("pluralize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestZeroValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"string", `""`},
		{"int", "0"},
		{"int32", "0"},
		{"int64", "0"},
		{"uint", "0"},
		{"uint32", "0"},
		{"uint64", "0"},
		{"float32", "0.0"},
		{"float64", "0.0"},
		{"bool", "false"},
		{"*Customer", "nil"},
		{"[]string", "nil"},
		{"map[string]int", "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := zeroValue(tt.input)
			if got != tt.want {
				t.Errorf("zeroValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToTableName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Customer", "customers"},
		{"OrderItem", "order_items"},
		{"Entity", "entities"},
		{"", "s"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toTableName(tt.input)
			if got != tt.want {
				t.Errorf("toTableName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToColumnName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"CustomerID", "customer_i_d"},
		{"FirstName", "first_name"},
		{"Email", "email"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toColumnName(tt.input)
			if got != tt.want {
				t.Errorf("toColumnName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTemplateEngine_ExecuteWithHelperFunctions(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine() failed: %v", err)
	}

	// Test that helper functions work in templates
	tests := []struct {
		name         string
		templateName string
		data         interface{}
		wantContains []string
	}{
		{
			name:         "usecase uses toCamelCase for dependencies",
			templateName: "domain/usecase/usecase.go.tmpl",
			data: UseCaseTemplateData{
				PackageName: "domain",
				UseCaseName: "CreateCustomer",
				Dependencies: []Dependency{
					{Name: "customerRepo", Type: "CustomerRepository"},
				},
				ModulePath: "github.com/test/project",
			},
			wantContains: []string{
				"customerRepo domain.CustomerRepository",
			},
		},
		{
			name:         "repository uses correct naming",
			templateName: "domain/entity/repository.go.tmpl",
			data: EntityTemplateData{
				PackageName:    "domain",
				EntityName:     "Customer",
				RepositoryName: "CustomerRepository",
				ModulePath:     "github.com/test/project",
			},
			wantContains: []string{
				"CustomerRepository",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Execute(tt.templateName, tt.data)
			if err != nil {
				t.Fatalf("Execute() failed: %v", err)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("Execute() result does not contain %q\nGot:\n%s", want, result)
				}
			}
		})
	}
}

func TestTemplateEngine_LoadsAllTemplates(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine() failed: %v", err)
	}

	// Verify that key templates are loaded under their layered paths
	requiredTemplates := []string{
		"domain/entity/entity.go.tmpl",
		"domain/entity/repository.go.tmpl",
		"domain/entity/entity_test.go.tmpl",
		"domain/usecase/usecase.go.tmpl",
		"domain/usecase/usecase_module.go.tmpl",
		"domain/usecase/usecase_test.go.tmpl",
		"ep/http/handler/http_handler.go.tmpl",
		"ep/http/handler/health_handler.go.tmpl",
		"ep/http/handler/module.go.tmpl",
		"ep/http/handler_test.go.tmpl",
		"ep/http/server.go.tmpl",
		"ep/http/middleware/cors.go.tmpl",
		"ep/http/routes/routes.go.tmpl",
		"ep/http/routes/router.go.tmpl",
		"ep/http/routes/health_router.go.tmpl",
		"ep/consumer/component_module.go.tmpl",
		"ep/consumer/rabbitmq/event_consumer.go.tmpl",
		"ep/consumer/rabbitmq/command_consumer.go.tmpl",
		"ep/consumer/rabbitmq/query_consumer.go.tmpl",
		"ep/consumer/rabbitmq/notification_consumer.go.tmpl",
		"ep/consumer/ibmmq/listener_consumer.go.tmpl",
		"ep/consumer/ibmmq/requestreply_consumer.go.tmpl",
		"da/database/postgres_adapter.go.tmpl",
		"da/database/postgres_repository.go.tmpl",
		"da/database/postgres_repository_test.go.tmpl",
		"da/database/mongodb_adapter.go.tmpl",
		"da/database/mongodb_repository.go.tmpl",
		"da/database/mongodb_repository_test.go.tmpl",
		"da/database/dynamodb_adapter.go.tmpl",
		"da/database/dynamodb_repository.go.tmpl",
		"da/database/dynamodb_repository_test.go.tmpl",
		"da/database/database_module.go.tmpl",
		"da/api/api_client.go.tmpl",
		"da/api/api_module.go.tmpl",
		"da/publisher/component_module.go.tmpl",
		"da/publisher/rabbitmq/event_publisher.go.tmpl",
		"da/publisher/rabbitmq/command_publisher.go.tmpl",
		"da/publisher/rabbitmq/query_publisher.go.tmpl",
		"da/publisher/rabbitmq/notification_publisher.go.tmpl",
		"da/publisher/ibmmq/sender_publisher.go.tmpl",
		"da/publisher/ibmmq/requestreply_publisher.go.tmpl",
		"infrastructure/messaging/rabbitmq/module.go.tmpl",
		"infrastructure/messaging/ibmmq/module.go.tmpl",
		"da/cache/cache_adapter.go.tmpl",
		"da/cache/cache_module.go.tmpl",
		"project/main.go.tmpl",
		"project/go.mod.tmpl",
		"project/config.go.tmpl",
		"project/config.yaml.tmpl",
		"project/README.md.tmpl",
		"project/Makefile.tmpl",
	}

	for _, tmplName := range requiredTemplates {
		t.Run(tmplName, func(t *testing.T) {
			tmpl := engine.templates.Lookup(tmplName)
			if tmpl == nil {
				t.Errorf("Template %q not loaded", tmplName)
			}
		})
	}
}

func TestTemplateEngine_ExecuteWithInvalidData(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine() failed: %v", err)
	}

	tests := []struct {
		name         string
		templateName string
		data         interface{}
		wantErr      bool
	}{
		{
			name:         "wrong data type",
			templateName: "domain/entity/entity.go.tmpl",
			data:         "invalid string data",
			wantErr:      true,
		},
		{
			name:         "missing required fields",
			templateName: "domain/entity/entity.go.tmpl",
			data:         EntityTemplateData{PackageName: "domain"}, // Minimal data should work
			wantErr:      false,                                     // Templates handle empty data gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Execute(tt.templateName, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateEngine_ExecuteMultipleTemplates(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine() failed: %v", err)
	}

	// Test executing multiple templates in sequence
	templates := []struct {
		name string
		data interface{}
	}{
		{
			name: "domain/entity/entity.go.tmpl",
			data: EntityTemplateData{
				PackageName: "domain",
				EntityName:  "Product",
				Fields:      []Field{{Name: "ID", Type: "string"}},
				ModulePath:  "github.com/test/project",
			},
		},
		{
			name: "domain/entity/repository.go.tmpl",
			data: EntityTemplateData{
				PackageName:    "domain",
				EntityName:     "Product",
				RepositoryName: "ProductRepository",
				ModulePath:     "github.com/test/project",
			},
		},
		{
			name: "domain/usecase/usecase.go.tmpl",
			data: UseCaseTemplateData{
				PackageName: "domain",
				UseCaseName: "CreateProduct",
				ModulePath:  "github.com/test/project",
			},
		},
	}

	for _, tmpl := range templates {
		t.Run(tmpl.name, func(t *testing.T) {
			result, err := engine.Execute(tmpl.name, tmpl.data)
			if err != nil {
				t.Errorf("Execute() failed for %s: %v", tmpl.name, err)
			}
			if result == "" {
				t.Errorf("Execute() returned empty result for %s", tmpl.name)
			}
		})
	}
}

func TestHelperFunctions_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		function func(string) string
		input    string
		want     string
	}{
		// toSnakeCase edge cases
		{"toSnakeCase with single char", toSnakeCase, "A", "a"},
		{"toSnakeCase with numbers", toSnakeCase, "User123", "user123"},
		{"toSnakeCase already snake", toSnakeCase, "already_snake", "already_snake"},

		// toCamelCase edge cases
		{"toCamelCase single char", toCamelCase, "A", "a"},
		{"toCamelCase already camel", toCamelCase, "alreadyCamel", "alreadyCamel"},

		// toKebabCase edge cases
		{"toKebabCase single char", toKebabCase, "A", "a"},
		{"toKebabCase with numbers", toKebabCase, "User123", "user123"},

		// pluralize edge cases
		{"pluralize already plural", pluralize, "customers", "customerses"}, // "customers" ends in "s", so adds "es"
		{"pluralize ending in y", pluralize, "category", "categories"},
		{"pluralize ending in s", pluralize, "bus", "buses"},
		{"pluralize single char", pluralize, "a", "as"},

		// toTableName edge cases
		{"toTableName single char", toTableName, "A", "as"},
		{"toTableName with numbers", toTableName, "User123", "user123s"},

		// toColumnName edge cases
		{"toColumnName single char", toColumnName, "A", "a"},
		{"toColumnName with numbers", toColumnName, "Field123", "field123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.function(tt.input)
			if got != tt.want {
				t.Errorf("%s(%q) = %q, want %q", tt.name, tt.input, got, tt.want)
			}
		})
	}
}

func TestZeroValue_AllTypes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Numeric types
		{"int8", "nil"},
		{"int16", "nil"},
		{"uint8", "nil"},
		{"uint16", "nil"},
		{"byte", "nil"},
		{"rune", "nil"},

		// Complex types
		{"complex64", "nil"},
		{"complex128", "nil"},

		// Pointer types
		{"*int", "nil"},
		{"*string", "nil"},

		// Slice types
		{"[]int", "nil"},
		{"[]byte", "nil"},

		// Map types
		{"map[string]string", "nil"},
		{"map[int]bool", "nil"},

		// Channel types
		{"chan int", "nil"},
		{"chan<- string", "nil"},
		{"<-chan bool", "nil"},

		// Interface types
		{"interface{}", "nil"},
		{"error", "nil"},

		// Custom types
		{"CustomType", "nil"},
		{"pkg.Type", "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := zeroValue(tt.input)
			if got != tt.want {
				t.Errorf("zeroValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTemplateEngine_ConcurrentExecution(t *testing.T) {
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine() failed: %v", err)
	}

	// Test concurrent template execution
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			data := EntityTemplateData{
				PackageName: "domain",
				EntityName:  "Customer",
				Fields:      []Field{{Name: "ID", Type: "string"}},
				ModulePath:  "github.com/test/project",
			}
			_, err := engine.Execute("domain/entity/entity.go.tmpl", data)
			if err != nil {
				t.Errorf("Concurrent execution %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
