package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "validation error with value",
			err: &ValidationError{
				Field:   "name",
				Value:   "invalid-name",
				Message: "must start with uppercase letter",
			},
			expected: "validation failed for name 'invalid-name': must start with uppercase letter",
		},
		{
			name: "validation error without value",
			err: &ValidationError{
				Field:   "fields",
				Value:   "",
				Message: "cannot be empty",
			},
			expected: "validation failed for fields: cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFileSystemError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *FileSystemError
		contains []string
	}{
		{
			name: "file system error with underlying error",
			err: &FileSystemError{
				Operation: "write",
				Path:      "/path/to/file.go",
				Err:       errors.New("permission denied"),
			},
			contains: []string{"file system error", "write", "/path/to/file.go", "permission denied"},
		},
		{
			name: "file system error without underlying error",
			err: &FileSystemError{
				Operation: "create",
				Path:      "/path/to/dir",
				Err:       nil,
			},
			contains: []string{"file system error", "create", "/path/to/dir"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("FileSystemError.Error() = %v, should contain %v", got, substr)
				}
			}
		})
	}
}

func TestFileSystemError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := &FileSystemError{
		Operation: "write",
		Path:      "/path/to/file.go",
		Err:       underlyingErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("FileSystemError.Unwrap() = %v, want %v", unwrapped, underlyingErr)
	}
}

func TestTemplateError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *TemplateError
		contains []string
	}{
		{
			name: "template error with underlying error",
			err: &TemplateError{
				TemplateName: "entity.go.tmpl",
				Operation:    "execution",
				Err:          errors.New("template not found"),
			},
			contains: []string{"template error", "entity.go.tmpl", "execution", "template not found"},
		},
		{
			name: "template error without underlying error",
			err: &TemplateError{
				TemplateName: "handler.go.tmpl",
				Operation:    "parsing",
				Err:          nil,
			},
			contains: []string{"template error", "handler.go.tmpl", "parsing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("TemplateError.Error() = %v, should contain %v", got, substr)
				}
			}
		})
	}
}

func TestTemplateError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := &TemplateError{
		TemplateName: "entity.go.tmpl",
		Operation:    "execution",
		Err:          underlyingErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("TemplateError.Unwrap() = %v, want %v", unwrapped, underlyingErr)
	}
}

func TestGenerationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *GenerationError
		contains []string
	}{
		{
			name: "generation error with underlying error",
			err: &GenerationError{
				Component: "Entity",
				Stage:     "validation",
				Err:       errors.New("invalid field type"),
			},
			contains: []string{"generation failed", "Entity", "validation", "invalid field type"},
		},
		{
			name: "generation error without underlying error",
			err: &GenerationError{
				Component: "Handler",
				Stage:     "template processing",
				Err:       nil,
			},
			contains: []string{"generation failed", "Handler", "template processing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("GenerationError.Error() = %v, should contain %v", got, substr)
				}
			}
		})
	}
}

func TestGenerationError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := &GenerationError{
		Component: "Entity",
		Stage:     "validation",
		Err:       underlyingErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("GenerationError.Unwrap() = %v, want %v", unwrapped, underlyingErr)
	}
}

func TestErrorTypes_ImplementErrorInterface(t *testing.T) {
	// Verify all error types implement the error interface
	var _ error = &ValidationError{}
	var _ error = &FileSystemError{}
	var _ error = &TemplateError{}
	var _ error = &GenerationError{}
}
