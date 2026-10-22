package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/template"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// seedDomainForProperty writes minimal domain files so DatabaseAdapterGenerator.Validate()
// passes inside property-test closures (which can't use require/*testing.T helpers).
func seedDomainForProperty(projectDir, entityName string) bool {
	domainDir := filepath.Join(projectDir, "internal", "domain")
	if err := os.MkdirAll(domainDir, 0o755); err != nil {
		return false
	}
	lower := strings.ToLower(entityName)
	if err := os.WriteFile(filepath.Join(domainDir, lower+".go"), []byte("package domain\n"), 0o644); err != nil {
		return false
	}
	if err := os.WriteFile(filepath.Join(domainDir, lower+"_repository.go"), []byte("package domain\n"), 0o644); err != nil {
		return false
	}
	return true
}

// genDatabaseType generates valid database types
func genDatabaseType() gopter.Gen {
	return gen.OneConstOf("postgres", "mongodb", "dynamodb")
}

// genEntityName generates valid entity names
func genEntityName() gopter.Gen {
	return gen.RegexMatch(`^[A-Z][a-zA-Z0-9]{0,20}`)
}

// findAdapterContent locates the engine-named adapter file for the given database type.
func findAdapterContent(files []GeneratedFile, dbType string) string {
	target := dbType + "_adapter.go"
	for _, f := range files {
		if strings.HasSuffix(f.Path, target) {
			return f.Content
		}
	}
	return ""
}

// findRepositoryContent locates the entity-named repository file.
func findRepositoryContent(files []GeneratedFile, entityName string) string {
	target := strings.ToLower(entityName) + "_repository.go"
	for _, f := range files {
		if strings.HasSuffix(f.Path, target) {
			return f.Content
		}
	}
	return ""
}

// Feature: go-microservice-scaffolding-cli, Property 29: Database adapter generation varies by database type
// **Validates: Requirements 9.2, 9.3, 9.4**
func TestProperty_DatabaseAdapterGenerationVariesByDatabaseType(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("postgres repository contains SQL queries", prop.ForAll(
		func(entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator("postgres", entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			repoContent := findRepositoryContent(files, entityName)
			if repoContent == "" {
				t.Logf("repository file not found")
				return false
			}

			sqlKeywords := []string{"INSERT INTO", "SELECT", "UPDATE", "DELETE FROM", "WHERE"}
			for _, keyword := range sqlKeywords {
				if !strings.Contains(repoContent, keyword) {
					t.Logf("SQL keyword %s not found in postgres repository", keyword)
					return false
				}
			}

			adapterContent := findAdapterContent(files, "postgres")
			if !strings.Contains(adapterContent, "pgxpool") {
				t.Logf("pgxpool not found in postgres adapter")
				return false
			}

			return true
		},
		genEntityName(),
		genFieldSpec(),
	))

	properties.Property("mongodb repository contains MongoDB queries", prop.ForAll(
		func(entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator("mongodb", entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			repoContent := findRepositoryContent(files, entityName)
			if repoContent == "" {
				t.Logf("repository file not found")
				return false
			}

			mongoOperations := []string{"InsertOne", "FindOne", "UpdateOne", "DeleteOne", "bson.M"}
			for _, operation := range mongoOperations {
				if !strings.Contains(repoContent, operation) {
					t.Logf("MongoDB operation %s not found in mongodb repository", operation)
					return false
				}
			}

			adapterContent := findAdapterContent(files, "mongodb")
			if !strings.Contains(adapterContent, "*mongo.Client") {
				t.Logf("*mongo.Client not found in mongodb adapter")
				return false
			}

			return true
		},
		genEntityName(),
		genFieldSpec(),
	))

	properties.Property("dynamodb repository contains DynamoDB queries", prop.ForAll(
		func(entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator("dynamodb", entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			repoContent := findRepositoryContent(files, entityName)
			if repoContent == "" {
				t.Logf("repository file not found")
				return false
			}

			dynamoOperations := []string{"PutItem", "GetItem", "DeleteItem", "Scan", "attributevalue"}
			for _, operation := range dynamoOperations {
				if !strings.Contains(repoContent, operation) {
					t.Logf("DynamoDB operation %s not found in dynamodb repository", operation)
					return false
				}
			}

			adapterContent := findAdapterContent(files, "dynamodb")
			if !strings.Contains(adapterContent, "*dynamodb.Client") {
				t.Logf("*dynamodb.Client not found in dynamodb adapter")
				return false
			}

			return true
		},
		genEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 30: Database adapter includes database connection in constructor
// **Validates: Requirements 9.5**
func TestProperty_DatabaseAdapterIncludesDatabaseConnectionInConstructor(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated adapter constructor includes database connection resource", prop.ForAll(
		func(databaseType string, entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator(databaseType, entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			adapterContent := findAdapterContent(files, databaseType)
			if adapterContent == "" {
				t.Logf("adapter file not found")
				return false
			}

			var expectedConstructor string
			var expectedResource string
			switch databaseType {
			case "postgres":
				expectedConstructor = "func NewPostgresAdapter"
				expectedResource = "pgxpool.Pool"
			case "mongodb":
				expectedConstructor = "func NewMongoDBAdapter"
				expectedResource = "*mongo.Client"
			case "dynamodb":
				expectedConstructor = "func NewDynamoDBAdapter"
				expectedResource = "*dynamodb.Client"
			}

			if !strings.Contains(adapterContent, expectedConstructor) {
				t.Logf("constructor %s not found", expectedConstructor)
				return false
			}
			if !strings.Contains(adapterContent, expectedResource) {
				t.Logf("database resource %s not found in adapter", expectedResource)
				return false
			}

			return true
		},
		genDatabaseType(),
		genEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 23: Entrypoints and adapters include logger in constructor
// **Validates: Requirements 5.5, 7.5, 8.5, 10.5, 11.5, 14.1**
func TestProperty_DatabaseAdapterIncludesLoggerInConstructor(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated adapter constructor includes Zap logger parameter", prop.ForAll(
		func(databaseType string, entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator(databaseType, entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			adapterContent := findAdapterContent(files, databaseType)
			if adapterContent == "" {
				t.Logf("adapter file not found")
				return false
			}

			if !strings.Contains(adapterContent, "logger") || !strings.Contains(adapterContent, "*zap.Logger") {
				t.Logf("logger not found in adapter")
				return false
			}

			return true
		},
		genDatabaseType(),
		genEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 31: Adapters implement domain interfaces
// **Validates: Requirements 9.8, 10.7**
func TestProperty_DatabaseAdapterImplementsDomainInterface(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated repository implements repository interface methods", prop.ForAll(
		func(databaseType string, entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator(databaseType, entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			repoContent := findRepositoryContent(files, entityName)
			if repoContent == "" {
				t.Logf("repository file not found")
				return false
			}

			requiredMethods := []string{
				"Create(ctx context.Context, entity *domain." + entityName + ") error",
				"GetByID(ctx context.Context, id string) (*domain." + entityName + ", error)",
				"Update(ctx context.Context, entity *domain." + entityName + ") error",
				"Delete(ctx context.Context, id string) error",
				"List(ctx context.Context, limit, offset int) ([]*domain." + entityName + ", error)",
			}

			for _, method := range requiredMethods {
				if !strings.Contains(repoContent, method) {
					t.Logf("required method not found: %s", method)
					return false
				}
			}

			return true
		},
		genDatabaseType(),
		genEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 53: Generated entrypoints and adapters are wired as fx Modules
// **Validates: Requirements 15.6**
func TestProperty_DatabaseAdapterWiredAsFxModule(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated adapter includes fx Module", prop.ForAll(
		func(databaseType string, entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator(databaseType, entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

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

			if !strings.Contains(moduleContent, "var Module = fx.Module") {
				t.Logf("fx.Module definition not found")
				return false
			}

			if !strings.Contains(moduleContent, "fx.Provide") {
				t.Logf("fx.Provide not found")
				return false
			}

			constructorName := "New" + entityName + "Repository"
			if !strings.Contains(moduleContent, constructorName) {
				t.Logf("constructor %s not found in module", constructorName)
				return false
			}

			return true
		},
		genDatabaseType(),
		genEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 7: Component generation creates files in correct directories
// **Validates: Requirements 2.1, 3.1, 4.1, 5.1, 7.1, 8.1, 9.1, 10.1, 11.1, 12.1**
func TestProperty_DatabaseAdapterCreatesFilesInCorrectDirectory(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated adapter files are in drivenadapters/database directory", prop.ForAll(
		func(databaseType string, entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator(databaseType, entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			expectedDir := filepath.Join("internal", "drivenadapters", "database")
			for _, file := range files {
				if !strings.Contains(file.Path, expectedDir) {
					t.Logf("file not in correct directory: %s", file.Path)
					return false
				}
			}

			return true
		},
		genDatabaseType(),
		genEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 6: Component name validation rejects invalid characters
// **Validates: Requirements 1.6, 2.7, 3.4, 4.3, 5.6, 7.6, 8.6, 9.7, 10.6, 11.6, 12.6**
func TestProperty_DatabaseAdapterNameValidationRejectsInvalidCharacters(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	genInvalidName := gen.OneConstOf(
		"123invalid",
		"invalid-name!",
		"invalid name",
		"",
		"invalid-Name",
		"invalid_Name",
	)

	properties.Property("validation rejects invalid entity names", prop.ForAll(
		func(invalidName string) bool {
			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator("postgres", invalidName, nil, "", templates)
			err = g.Validate()

			if err == nil {
				t.Logf("expected error for invalid name %s, but got none", invalidName)
				return false
			}

			return true
		},
		genInvalidName,
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 28: Adapters include error handling and retry logic
// **Validates: Requirements 8.4, 9.6, 10.4, 11.4**
func TestProperty_DatabaseAdapterIncludesErrorHandling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated repository includes error handling", prop.ForAll(
		func(databaseType string, entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator(databaseType, entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			repoContent := findRepositoryContent(files, entityName)
			if repoContent == "" {
				t.Logf("repository file not found")
				return false
			}

			errorPatterns := []string{
				"if err != nil",
				".logger.Error",
				"return",
				"fmt.Errorf",
			}

			for _, pattern := range errorPatterns {
				if !strings.Contains(repoContent, pattern) {
					t.Logf("error handling pattern %s not found", pattern)
					return false
				}
			}

			return true
		},
		genDatabaseType(),
		genEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}

// Feature: go-microservice-scaffolding-cli, Property 54: Code generation uses text/template
// **Validates: Requirements 21.1, 21.2, 21.3**
func TestProperty_DatabaseAdapterUsesTextTemplate(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("adapter generated from templates has expected structure", prop.ForAll(
		func(databaseType string, entityName string, fields []string) bool {
			tmpDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			if err != nil {
				t.Logf("failed to create template engine: %v", err)
				return false
			}

			g := NewDatabaseAdapterGenerator(databaseType, entityName, fields, tmpDir, templates)

			if !seedDomainForProperty(tmpDir, entityName) {
				t.Logf("failed to seed domain for entity %q", entityName)
				return false
			}

			if err := g.Validate(); err != nil {
				t.Logf("validation failed: %v", err)
				return false
			}

			files, err := g.Generate()
			if err != nil {
				t.Logf("generation failed: %v", err)
				return false
			}

			adapterContent := findAdapterContent(files, databaseType)
			if adapterContent == "" {
				t.Logf("adapter file not found")
				return false
			}

			hasPackage := strings.Contains(adapterContent, "package database")
			hasType := strings.Contains(adapterContent, "type ")
			hasStruct := strings.Contains(adapterContent, "struct")

			if !hasPackage || !hasType || !hasStruct {
				t.Logf("missing expected template structure")
				return false
			}

			return true
		},
		genDatabaseType(),
		genEntityName(),
		genFieldSpec(),
	))

	properties.TestingRun(t)
}
