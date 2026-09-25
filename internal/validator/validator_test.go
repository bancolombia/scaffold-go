package validator

import (
	"os"
	"testing"
)

func TestValidateName(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "User", false},
		{"valid camelCase", "UserProfile", false},
		{"empty name", "", true},
		{"lowercase start", "user", true},
		{"with underscore", "User_Name", true},
		{"with hyphen", "User-Name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "my-project", false},
		{"valid simple", "project", false},
		{"empty name", "", true},
		{"uppercase start", "MyProject", true},
		{"with underscore", "my_project", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFields(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		input   []string
		wantErr bool
	}{
		{"valid fields", []string{"Name:string", "Age:int"}, false},
		{"invalid format", []string{"Name"}, true},
		{"empty name", []string{":string"}, true},
		{"empty type", []string{"Name:"}, true},
		{"lowercase field name", []string{"name:string"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateFields(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDirectoryNotExists(t *testing.T) {
	v := New()

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	existingDir := tmpDir + "/existing"
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"non-existent directory", tmpDir + "/nonexistent", false},
		{"existing directory", existingDir, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateDirectoryNotExists(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectoryNotExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDirectoryExists(t *testing.T) {
	v := New()

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	existingDir := tmpDir + "/existing"
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a file (not a directory)
	filePath := tmpDir + "/file.txt"
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"existing directory", existingDir, false},
		{"non-existent directory", tmpDir + "/nonexistent", true},
		{"file not directory", filePath, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateDirectoryExists(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectoryExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProjectInitialized(t *testing.T) {
	v := New()

	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create a project with go.mod in root
	projectRoot := tmpDir + "/project"
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatalf("Failed to create project root: %v", err)
	}
	if err := os.WriteFile(projectRoot+"/go.mod", []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a subdirectory within the project
	subDir := projectRoot + "/internal/domain"
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a directory without go.mod
	noGoModDir := tmpDir + "/nogomod"
	if err := os.MkdirAll(noGoModDir, 0755); err != nil {
		t.Fatalf("Failed to create nogomod directory: %v", err)
	}

	tests := []struct {
		name    string
		dir     string
		wantErr bool
	}{
		{"project root with go.mod", projectRoot, false},
		{"subdirectory of project", subDir, false},
		{"directory without go.mod", noGoModDir, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateProjectInitialized(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectInitialized() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
