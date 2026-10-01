package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Custom generators for cache property-based testing

// genValidCacheName generates valid cache names (PascalCase)
func genValidCacheName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{2,20}`)
}

// Feature: go-microservice-scaffolding-cli, Property 36: Cache adapter includes standard cache methods
// **Validates: Requirements 12.2**
func TestProperty_CacheAdapterIncludesStandardCacheMethods(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated cache adapter contains Get, Set, Delete, and Exists methods", prop.ForAll(
		func(cacheName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewCacheGenerator(cacheName, tmpDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find cache file (not module)
			var cacheContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "_cache.go") {
					cacheContent = file.Content
					break
				}
			}

			if cacheContent == "" {
				t.Logf("cache file not found")
				return false
			}

			// Check for Get method signature
			getMethodSignature := "func (c *" + cacheName + "Cache) Get(ctx context.Context, key string, dest interface{}) error"
			if !strings.Contains(cacheContent, getMethodSignature) {
				t.Logf("Get method not found")
				return false
			}

			// Check for Set method signature
			setMethodSignature := "func (c *" + cacheName + "Cache) Set(ctx context.Context, key string, value interface{}) error"
			if !strings.Contains(cacheContent, setMethodSignature) {
				t.Logf("Set method not found")
				return false
			}

			// Check for Delete method signature
			deleteMethodSignature := "func (c *" + cacheName + "Cache) Delete(ctx context.Context, key string) error"
			if !strings.Contains(cacheContent, deleteMethodSignature) {
				t.Logf("Delete method not found")
				return false
			}

			// Check for Exists method signature
			existsMethodSignature := "func (c *" + cacheName + "Cache) Exists(ctx context.Context, key string) (bool, error)"
			if !strings.Contains(cacheContent, existsMethodSignature) {
				t.Logf("Exists method not found")
				return false
			}

			// Verify all methods have context parameter
			if !strings.Contains(cacheContent, "ctx context.Context") {
				t.Logf("context parameter not found in methods")
				return false
			}

			// Verify key parameter in all methods
			keyParamCount := strings.Count(cacheContent, "key string")
			if keyParamCount < 4 {
				t.Logf("key parameter not found in all methods")
				return false
			}

			// Verify Get method has dest parameter for deserialization
			if !strings.Contains(cacheContent, "dest interface{}") {
				t.Logf("dest parameter not found in Get method")
				return false
			}

			// Verify Set method has value parameter
			if !strings.Contains(cacheContent, "value interface{}") {
				t.Logf("value parameter not found in Set method")
				return false
			}

			// Verify Exists method returns bool
			if !strings.Contains(cacheContent, "Exists(ctx context.Context, key string) (bool, error)") {
				t.Logf("Exists method does not return bool")
				return false
			}

			return true
		},
		genValidCacheName(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 37: Cache adapter constructor includes cache connection
// **Validates: Requirements 12.3**
func TestProperty_CacheAdapterConstructorIncludesCacheConnection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated cache adapter constructor includes cache connection placeholder", prop.ForAll(
		func(cacheName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewCacheGenerator(cacheName, tmpDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find cache file
			var cacheContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "_cache.go") {
					cacheContent = file.Content
					break
				}
			}

			if cacheContent == "" {
				t.Logf("cache file not found")
				return false
			}

			// Check for constructor signature
			constructorSignature := "func New" + cacheName + "Cache(logger *zap.Logger)"
			if !strings.Contains(cacheContent, constructorSignature) {
				t.Logf("constructor signature not found")
				return false
			}

			// Check for cache connection placeholder in struct
			if !strings.Contains(cacheContent, "TODO: Add cache connection") {
				t.Logf("cache connection placeholder not found in struct")
				return false
			}

			// Verify the TODO comment mentions specific cache types
			if !strings.Contains(cacheContent, "*redis.Client") {
				t.Logf("redis.Client example not found")
				return false
			}

			// Check for logger field in struct
			if !strings.Contains(cacheContent, "logger *zap.Logger") {
				t.Logf("logger field not found in struct")
				return false
			}

			// Verify constructor initializes the struct
			if !strings.Contains(cacheContent, "return &"+cacheName+"Cache{") {
				t.Logf("constructor does not initialize struct")
				return false
			}

			// Verify logger is initialized in constructor
			if !strings.Contains(cacheContent, "logger: logger,") {
				t.Logf("logger not initialized in constructor")
				return false
			}

			return true
		},
		genValidCacheName(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 38: Cache adapter includes TTL support
// **Validates: Requirements 12.4**
func TestProperty_CacheAdapterIncludesTTLSupport(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated cache adapter includes TTL configuration support", prop.ForAll(
		func(cacheName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewCacheGenerator(cacheName, tmpDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find cache file
			var cacheContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "_cache.go") {
					cacheContent = file.Content
					break
				}
			}

			if cacheContent == "" {
				t.Logf("cache file not found")
				return false
			}

			// Check for TTL field in struct
			if !strings.Contains(cacheContent, "ttl time.Duration") {
				t.Logf("ttl field not found in struct")
				return false
			}

			// Check for SetTTL method
			setTTLMethodSignature := "func (c *" + cacheName + "Cache) SetTTL(ttl time.Duration)"
			if !strings.Contains(cacheContent, setTTLMethodSignature) {
				t.Logf("SetTTL method not found")
				return false
			}

			// Verify TTL is initialized in constructor with default value
			if !strings.Contains(cacheContent, "ttl:") {
				t.Logf("ttl not initialized in constructor")
				return false
			}

			// Verify default TTL value is set
			if !strings.Contains(cacheContent, "5 * time.Minute") {
				t.Logf("default TTL value not found")
				return false
			}

			// Verify time package is imported
			if !strings.Contains(cacheContent, "time") {
				t.Logf("time package not imported")
				return false
			}

			// Verify TTL is used in Set method (in TODO comment)
			if !strings.Contains(cacheContent, "c.ttl") {
				t.Logf("TTL not referenced in Set method")
				return false
			}

			// Verify SetTTL method updates the ttl field
			if !strings.Contains(cacheContent, "c.ttl = ttl") {
				t.Logf("SetTTL method does not update ttl field")
				return false
			}

			return true
		},
		genValidCacheName(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 39: Cache adapter includes error handling
// **Validates: Requirements 12.5**
func TestProperty_CacheAdapterIncludesErrorHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated cache adapter includes error handling logic", prop.ForAll(
		func(cacheName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewCacheGenerator(cacheName, tmpDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find cache file
			var cacheContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "_cache.go") {
					cacheContent = file.Content
					break
				}
			}

			if cacheContent == "" {
				t.Logf("cache file not found")
				return false
			}

			// Verify all methods return error
			if !strings.Contains(cacheContent, "Get(ctx context.Context, key string, dest interface{}) error") {
				t.Logf("Get method does not return error")
				return false
			}

			if !strings.Contains(cacheContent, "Set(ctx context.Context, key string, value interface{}) error") {
				t.Logf("Set method does not return error")
				return false
			}

			if !strings.Contains(cacheContent, "Delete(ctx context.Context, key string) error") {
				t.Logf("Delete method does not return error")
				return false
			}

			if !strings.Contains(cacheContent, "Exists(ctx context.Context, key string) (bool, error)") {
				t.Logf("Exists method does not return error")
				return false
			}

			// Verify error handling in Set method
			if !strings.Contains(cacheContent, "if err != nil") {
				t.Logf("error checking not found")
				return false
			}

			// Verify error logging
			if !strings.Contains(cacheContent, "logger.Error") {
				t.Logf("error logging not found")
				return false
			}

			// Verify error is returned after logging
			if !strings.Contains(cacheContent, "return err") {
				t.Logf("error not returned")
				return false
			}

			// Verify specific error message for marshal failure
			if !strings.Contains(cacheContent, "Failed to marshal value") {
				t.Logf("marshal error message not found")
				return false
			}

			// Verify zap.Error is used for structured error logging
			if !strings.Contains(cacheContent, "zap.Error(err)") {
				t.Logf("zap.Error not used for structured logging")
				return false
			}

			return true
		},
		genValidCacheName(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 53: Generated adapters are wired as fx Modules
// **Validates: Requirements 15.6**
func TestProperty_GeneratedCacheAdaptersAreWiredAsFxModules(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated cache includes fx Module definition", prop.ForAll(
		func(cacheName string) bool {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create template engine
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			// Create generator
			gen := NewCacheGenerator(cacheName, tmpDir, templates)

			// Validate
			if err := gen.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			// Generate
			files, err := gen.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			// Find module file
			var moduleContent string
			for _, file := range files {
				if strings.HasSuffix(file.Path, "module.go") {
					moduleContent = file.Content
					break
				}
			}

			if moduleContent == "" {
				t.Logf("module file not found")
				return false
			}

			// Verify module file contains fx import
			if !strings.Contains(moduleContent, "go.uber.org/fx") {
				t.Logf("module file does not import go.uber.org/fx")
				return false
			}

			// Verify module file contains Module variable
			if !strings.Contains(moduleContent, "var Module = fx.Module") {
				t.Logf("module file does not contain Module variable")
				return false
			}

			// Verify module file contains fx.Provide
			if !strings.Contains(moduleContent, "fx.Provide") {
				t.Logf("module file does not contain fx.Provide")
				return false
			}

			// Verify module provides the cache constructor
			constructorName := "New" + cacheName + "Cache"
			if !strings.Contains(moduleContent, constructorName) {
				t.Logf("module does not provide constructor %s", constructorName)
				return false
			}

			// Verify module is in cache package
			if !strings.Contains(moduleContent, "package cache") {
				t.Logf("module file not in cache package")
				return false
			}

			// Verify module name is "cache"
			if !strings.Contains(moduleContent, `fx.Module("cache"`) {
				t.Logf("module name is not 'cache'")
				return false
			}

			// Verify module file is in the correct directory
			for _, file := range files {
				if strings.HasSuffix(file.Path, "module.go") {
					if !strings.Contains(file.Path, filepath.Join("internal", "drivenadapters", "cache")) {
						t.Logf("module file not in drivenadapters/cache directory")
						return false
					}
				}
			}

			return true
		},
		genValidCacheName(),
	))

	properties.TestingRun(t)
}
