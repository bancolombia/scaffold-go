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

// seedDomainEntity writes minimal entity + repository files into
// <projectDir>/internal/domain so DatabaseAdapterGenerator.Validate() passes.
func seedDomainEntity(t *testing.T, projectDir, entityName string) {
	t.Helper()
	domainDir := filepath.Join(projectDir, "internal", "domain")
	require.NoError(t, os.MkdirAll(domainDir, 0o755))
	lower := strings.ToLower(entityName)
	require.NoError(t, os.WriteFile(filepath.Join(domainDir, lower+".go"), []byte("package domain\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(domainDir, lower+"_repository.go"), []byte("package domain\n"), 0o644))
}

// findFileBySuffix returns the first GeneratedFile whose Path ends with the given suffix.
func findFileBySuffix(t *testing.T, files []GeneratedFile, suffix string) GeneratedFile {
	t.Helper()
	for _, f := range files {
		if strings.HasSuffix(f.Path, suffix) {
			return f
		}
	}
	t.Fatalf("no generated file ends with %q; got: %v", suffix, filePaths(files))
	return GeneratedFile{}
}

func filePaths(files []GeneratedFile) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.Path)
	}
	return out
}

func TestDatabaseAdapterGenerator_Validate(t *testing.T) {
	tests := []struct {
		name         string
		entityName   string
		databaseType string
		wantErr      bool
	}{
		{
			name:         "valid postgres adapter",
			entityName:   "Customer",
			databaseType: "postgres",
			wantErr:      false,
		},
		{
			name:         "valid mongodb adapter",
			entityName:   "Product",
			databaseType: "mongodb",
			wantErr:      false,
		},
		{
			name:         "valid dynamodb adapter",
			entityName:   "Order",
			databaseType: "dynamodb",
			wantErr:      false,
		},
		{
			name:         "invalid entity name with special chars",
			entityName:   "Customer-!",
			databaseType: "postgres",
			wantErr:      true,
		},
		{
			name:         "invalid entity name starting with number",
			entityName:   "123Entity",
			databaseType: "postgres",
			wantErr:      true,
		},
		{
			name:         "empty entity name",
			entityName:   "",
			databaseType: "postgres",
			wantErr:      true,
		},
		{
			name:         "invalid database type",
			entityName:   "Customer",
			databaseType: "mysql",
			wantErr:      true,
		},
		{
			name:         "empty database type",
			entityName:   "Customer",
			databaseType: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			gen := NewDatabaseAdapterGenerator(tt.databaseType, tt.entityName, nil, "", templates)
			err = gen.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDatabaseAdapterGenerator_Generate_Postgres(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	fields := []string{"ID:string", "Name:string", "Email:string"}
	gen := NewDatabaseAdapterGenerator("postgres", "Customer", fields, tempDir, templates)

	files, err := gen.Generate()
	require.NoError(t, err)
	assert.Len(t, files, 4) // adapter + repository + repository_test + module

	adapterFile := findFileBySuffix(t, files, "postgres_adapter.go")
	assert.Equal(t, filepath.Join(tempDir, "internal", "drivenadapters", "database", "postgres_adapter.go"), adapterFile.Path)
	assert.Contains(t, adapterFile.Content, "type PostgresAdapter struct")
	assert.Contains(t, adapterFile.Content, "pgxpool.Pool")
	assert.Contains(t, adapterFile.Content, "func NewPostgresAdapter(cfg *config.Config, logger *zap.Logger)")

	repoFile := findFileBySuffix(t, files, "customer_repository.go")
	assert.Contains(t, repoFile.Content, "type CustomerRepository struct")
	assert.Contains(t, repoFile.Content, "func NewCustomerRepository")
	assert.Contains(t, repoFile.Content, "func (a *CustomerRepository) Create(ctx context.Context, entity *domain.Customer) error")
	assert.Contains(t, repoFile.Content, "func (a *CustomerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error)")
	assert.Contains(t, repoFile.Content, "func (a *CustomerRepository) Update(ctx context.Context, entity *domain.Customer) error")
	assert.Contains(t, repoFile.Content, "func (a *CustomerRepository) Delete(ctx context.Context, id string) error")
	assert.Contains(t, repoFile.Content, "func (a *CustomerRepository) List(ctx context.Context, limit, offset int) ([]*domain.Customer, error)")
	assert.Contains(t, repoFile.Content, "INSERT INTO")
	assert.Contains(t, repoFile.Content, "SELECT")
	assert.Contains(t, repoFile.Content, "UPDATE")
	assert.Contains(t, repoFile.Content, "DELETE FROM")

	testFile := findFileBySuffix(t, files, "customer_repository_test.go")
	assert.Contains(t, testFile.Content, "func TestCustomerRepository_Create")
	assert.Contains(t, testFile.Content, "pgxmock")

	moduleFile := findFileBySuffix(t, files, "module.go")
	assert.Equal(t, filepath.Join(tempDir, "internal", "drivenadapters", "database", "module.go"), moduleFile.Path)
	assert.Contains(t, moduleFile.Content, "package database")
	assert.Contains(t, moduleFile.Content, "fx.Module")
	assert.Contains(t, moduleFile.Content, "NewPostgresAdapter")
	assert.Contains(t, moduleFile.Content, "NewCustomerRepository")
}

func TestDatabaseAdapterGenerator_Generate_MongoDB(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	fields := []string{"ID:string", "Name:string"}
	gen := NewDatabaseAdapterGenerator("mongodb", "Product", fields, tempDir, templates)

	files, err := gen.Generate()
	require.NoError(t, err)
	assert.Len(t, files, 4)

	adapterFile := findFileBySuffix(t, files, "mongodb_adapter.go")
	assert.Contains(t, adapterFile.Content, "type MongoDBAdapter struct")
	assert.Contains(t, adapterFile.Content, "*mongo.Client")
	assert.Contains(t, adapterFile.Content, "func NewMongoDBAdapter(cfg *config.Config, logger *zap.Logger)")

	repoFile := findFileBySuffix(t, files, "product_repository.go")
	assert.Contains(t, repoFile.Content, "type ProductRepository struct")
	assert.Contains(t, repoFile.Content, "InsertOne")
	assert.Contains(t, repoFile.Content, "FindOne")
	assert.Contains(t, repoFile.Content, "UpdateOne")
	assert.Contains(t, repoFile.Content, "DeleteOne")
	assert.Contains(t, repoFile.Content, "bson.M")

	testFile := findFileBySuffix(t, files, "product_repository_test.go")
	assert.Contains(t, testFile.Content, "func TestProductRepository_Create")
	assert.Contains(t, testFile.Content, "mongo.InsertOneResult")
}

func TestDatabaseAdapterGenerator_Generate_DynamoDB(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	fields := []string{"ID:string", "Status:string"}
	gen := NewDatabaseAdapterGenerator("dynamodb", "Order", fields, tempDir, templates)

	files, err := gen.Generate()
	require.NoError(t, err)
	assert.Len(t, files, 4)

	adapterFile := findFileBySuffix(t, files, "dynamodb_adapter.go")
	assert.Contains(t, adapterFile.Content, "type DynamoDBAdapter struct")
	assert.Contains(t, adapterFile.Content, "*dynamodb.Client")
	assert.Contains(t, adapterFile.Content, "func NewDynamoDBAdapter(cfg *c.Config, logger *zap.Logger)")

	repoFile := findFileBySuffix(t, files, "order_repository.go")
	assert.Contains(t, repoFile.Content, "type OrderRepository struct")
	assert.Contains(t, repoFile.Content, "PutItem")
	assert.Contains(t, repoFile.Content, "GetItem")
	assert.Contains(t, repoFile.Content, "DeleteItem")
	assert.Contains(t, repoFile.Content, "Scan")
	assert.Contains(t, repoFile.Content, "attributevalue.MarshalMap")
}

func TestDatabaseAdapterGenerator_ComponentType(t *testing.T) {
	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewDatabaseAdapterGenerator("postgres", "Customer", nil, "", templates)
	assert.Equal(t, "database-adapter", gen.ComponentType())
}

func TestDatabaseAdapterGenerator_ParseFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want   int
	}{
		{
			name:   "parse multiple fields",
			fields: []string{"ID:string", "Name:string", "Email:string"},
			want:   3,
		},
		{
			name:   "parse single field",
			fields: []string{"ID:string"},
			want:   1,
		},
		{
			name:   "parse no fields",
			fields: []string{},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			gen := NewDatabaseAdapterGenerator("postgres", "Test", tt.fields, "", templates)
			parsed := gen.parseFields()

			assert.Len(t, parsed, tt.want)

			for i, field := range parsed {
				parts := strings.Split(tt.fields[i], ":")
				assert.Equal(t, strings.TrimSpace(parts[0]), field.Name)
				assert.Equal(t, strings.TrimSpace(parts[1]), field.Type)
			}
		})
	}
}

func TestDatabaseAdapterGenerator_Integration(t *testing.T) {
	tempDir := t.TempDir()
	seedDomainEntity(t, tempDir, "Customer")

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	fields := []string{"ID:string", "Name:string"}
	gen := NewDatabaseAdapterGenerator("postgres", "Customer", fields, tempDir, templates)

	err = gen.Validate()
	require.NoError(t, err)

	files, err := gen.Generate()
	require.NoError(t, err)

	for _, file := range files {
		dir := filepath.Dir(file.Path)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(file.Path, []byte(file.Content), 0644)
		require.NoError(t, err)
	}

	adapterPath := filepath.Join(tempDir, "internal", "drivenadapters", "database", "postgres_adapter.go")
	modulePath := filepath.Join(tempDir, "internal", "drivenadapters", "database", "module.go")

	assert.FileExists(t, adapterPath)
	assert.FileExists(t, modulePath)

	adapterContent, err := os.ReadFile(adapterPath)
	require.NoError(t, err)
	assert.Contains(t, string(adapterContent), "package database")

	moduleContent, err := os.ReadFile(modulePath)
	require.NoError(t, err)
	assert.Contains(t, string(moduleContent), "package database")
	assert.Contains(t, string(moduleContent), "fx.Module")
}

func TestDatabaseAdapterGenerator_LoggerIntegration(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewDatabaseAdapterGenerator("postgres", "Customer", nil, tempDir, templates)

	files, err := gen.Generate()
	require.NoError(t, err)

	adapterFile := findFileBySuffix(t, files, "postgres_adapter.go")

	assert.Contains(t, adapterFile.Content, "logger *zap.Logger")
	assert.Contains(t, adapterFile.Content, "logger: logger")

	repoFile := findFileBySuffix(t, files, "customer_repository.go")
	assert.Contains(t, repoFile.Content, "a.logger.Info")
	assert.Contains(t, repoFile.Content, "a.logger.Error")
	assert.Contains(t, repoFile.Content, "zap.String")
	assert.Contains(t, repoFile.Content, "zap.Error")
}

func TestDatabaseAdapterGenerator_DatabaseConnection(t *testing.T) {
	tests := []struct {
		name         string
		databaseType string
		adapterFile  string
		wantAdapter  string
		wantField    string
	}{
		{
			name:         "postgres has pgxpool",
			databaseType: "postgres",
			adapterFile:  "postgres_adapter.go",
			wantAdapter:  "type PostgresAdapter struct",
			wantField:    "pgxpool.Pool",
		},
		{
			name:         "mongodb has mongo.Client",
			databaseType: "mongodb",
			adapterFile:  "mongodb_adapter.go",
			wantAdapter:  "type MongoDBAdapter struct",
			wantField:    "*mongo.Client",
		},
		{
			name:         "dynamodb has dynamodb.Client",
			databaseType: "dynamodb",
			adapterFile:  "dynamodb_adapter.go",
			wantAdapter:  "type DynamoDBAdapter struct",
			wantField:    "*dynamodb.Client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			gen := NewDatabaseAdapterGenerator(tt.databaseType, "Test", nil, tempDir, templates)

			files, err := gen.Generate()
			require.NoError(t, err)

			adapterFile := findFileBySuffix(t, files, tt.adapterFile)
			assert.Contains(t, adapterFile.Content, tt.wantAdapter)
			assert.Contains(t, adapterFile.Content, tt.wantField)
		})
	}
}

func TestDatabaseAdapterGenerator_CRUDMethods(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewDatabaseAdapterGenerator("postgres", "Customer", nil, tempDir, templates)

	files, err := gen.Generate()
	require.NoError(t, err)

	repoFile := findFileBySuffix(t, files, "customer_repository.go")

	requiredMethods := []string{
		"func (a *CustomerRepository) Create(ctx context.Context, entity *domain.Customer) error",
		"func (a *CustomerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error)",
		"func (a *CustomerRepository) Update(ctx context.Context, entity *domain.Customer) error",
		"func (a *CustomerRepository) Delete(ctx context.Context, id string) error",
		"func (a *CustomerRepository) List(ctx context.Context, limit, offset int) ([]*domain.Customer, error)",
	}

	for _, method := range requiredMethods {
		assert.Contains(t, repoFile.Content, method, "Missing method: %s", method)
	}
}

func TestDatabaseAdapterGenerator_FxModuleStructure(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/demo\n\ngo 1.21\n"), 0o644))
	seedDomainEntity(t, tempDir, "Customer")

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewDatabaseAdapterGenerator("postgres", "Customer", nil, tempDir, templates)

	files, err := gen.Generate()
	require.NoError(t, err)

	moduleFile := findFileBySuffix(t, files, "module.go")

	assert.Contains(t, moduleFile.Content, "package database")
	assert.Contains(t, moduleFile.Content, `"go.uber.org/fx"`)
	assert.Contains(t, moduleFile.Content, `var Module = fx.Module("database"`)
	assert.Contains(t, moduleFile.Content, "fx.Provide(")
	assert.Contains(t, moduleFile.Content, "fx.Annotate(")
	assert.Contains(t, moduleFile.Content, "NewPostgresAdapter")
	assert.Contains(t, moduleFile.Content, "NewCustomerRepository")
	assert.Contains(t, moduleFile.Content, "fx.As(new(domain.ICustomerRepository))")
	assert.Contains(t, moduleFile.Content, `"example.com/demo/internal/domain"`)
}

func TestDatabaseAdapterGenerator_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	templates, err := template.NewTemplateEngine()
	require.NoError(t, err)

	gen := NewDatabaseAdapterGenerator("postgres", "Customer", nil, tempDir, templates)

	files, err := gen.Generate()
	require.NoError(t, err)

	repoFile := findFileBySuffix(t, files, "customer_repository.go")

	assert.Contains(t, repoFile.Content, "if err != nil")
	assert.Contains(t, repoFile.Content, "a.logger.Error")
	assert.Contains(t, repoFile.Content, "return")
	assert.Contains(t, repoFile.Content, "fmt.Errorf")
}

func TestDatabaseAdapterGenerator_CaseInsensitiveDatabaseType(t *testing.T) {
	tests := []struct {
		name         string
		databaseType string
		wantErr      bool
	}{
		{
			name:         "lowercase postgres",
			databaseType: "postgres",
			wantErr:      false,
		},
		{
			name:         "uppercase POSTGRES",
			databaseType: "POSTGRES",
			wantErr:      false,
		},
		{
			name:         "lowercase mongodb",
			databaseType: "mongodb",
			wantErr:      false,
		},
		{
			name:         "uppercase MONGODB",
			databaseType: "MONGODB",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := template.NewTemplateEngine()
			require.NoError(t, err)

			gen := NewDatabaseAdapterGenerator(tt.databaseType, "Test", nil, "", templates)
			err = gen.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
