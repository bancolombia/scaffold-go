package commands

import (
	"errors"
	"testing"

	"github.com/bancolombia/scaffold-go/internal/domain"
	"go.uber.org/zap"
)

func TestHandleError(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "nil error returns success",
			err:          nil,
			expectedCode: ExitCodeSuccess,
		},
		{
			name: "validation error returns validation exit code",
			err: &domain.ValidationError{
				Field:   "name",
				Value:   "invalid",
				Message: "must start with uppercase",
			},
			expectedCode: ExitCodeValidationError,
		},
		{
			name: "file system error returns file system exit code",
			err: &domain.FileSystemError{
				Operation: "write",
				Path:      "/path/to/file",
				Err:       errors.New("permission denied"),
			},
			expectedCode: ExitCodeFileSystemError,
		},
		{
			name: "template error returns template exit code",
			err: &domain.TemplateError{
				TemplateName: "entity.go.tmpl",
				Operation:    "execution",
				Err:          errors.New("template not found"),
			},
			expectedCode: ExitCodeTemplateError,
		},
		{
			name: "generation error returns generation exit code",
			err: &domain.GenerationError{
				Component: "Entity",
				Stage:     "validation",
				Err:       errors.New("invalid field"),
			},
			expectedCode: ExitCodeGenerationError,
		},
		{
			name:         "unknown error returns unknown exit code",
			err:          errors.New("some unknown error"),
			expectedCode: ExitCodeUnknownError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := HandleError(tt.err, logger)
			if code != tt.expectedCode {
				t.Errorf("HandleError() = %v, want %v", code, tt.expectedCode)
			}
		})
	}
}

func TestWrapValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		err      error
		wantType bool
	}{
		{
			name:     "wraps regular error as validation error",
			field:    "name",
			value:    "invalid",
			err:      errors.New("must start with uppercase"),
			wantType: true,
		},
		{
			name:  "returns existing validation error unchanged",
			field: "name",
			value: "invalid",
			err: &domain.ValidationError{
				Field:   "name",
				Value:   "invalid",
				Message: "must start with uppercase",
			},
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapValidationError(tt.field, tt.value, tt.err)
			var validationErr *domain.ValidationError
			if errors.As(result, &validationErr) != tt.wantType {
				t.Errorf("WrapValidationError() type assertion = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestWrapFileSystemError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		path      string
		err       error
		wantType  bool
	}{
		{
			name:      "wraps regular error as file system error",
			operation: "write",
			path:      "/path/to/file",
			err:       errors.New("permission denied"),
			wantType:  true,
		},
		{
			name:      "returns existing file system error unchanged",
			operation: "write",
			path:      "/path/to/file",
			err: &domain.FileSystemError{
				Operation: "write",
				Path:      "/path/to/file",
				Err:       errors.New("permission denied"),
			},
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapFileSystemError(tt.operation, tt.path, tt.err)
			var fileSystemErr *domain.FileSystemError
			if errors.As(result, &fileSystemErr) != tt.wantType {
				t.Errorf("WrapFileSystemError() type assertion = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestWrapTemplateError(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		operation    string
		err          error
		wantType     bool
	}{
		{
			name:         "wraps regular error as template error",
			templateName: "entity.go.tmpl",
			operation:    "execution",
			err:          errors.New("template not found"),
			wantType:     true,
		},
		{
			name:         "returns existing template error unchanged",
			templateName: "entity.go.tmpl",
			operation:    "execution",
			err: &domain.TemplateError{
				TemplateName: "entity.go.tmpl",
				Operation:    "execution",
				Err:          errors.New("template not found"),
			},
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapTemplateError(tt.templateName, tt.operation, tt.err)
			var templateErr *domain.TemplateError
			if errors.As(result, &templateErr) != tt.wantType {
				t.Errorf("WrapTemplateError() type assertion = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestWrapGenerationError(t *testing.T) {
	tests := []struct {
		name      string
		component string
		stage     string
		err       error
		wantType  bool
	}{
		{
			name:      "wraps regular error as generation error",
			component: "Entity",
			stage:     "validation",
			err:       errors.New("invalid field"),
			wantType:  true,
		},
		{
			name:      "returns existing generation error unchanged",
			component: "Entity",
			stage:     "validation",
			err: &domain.GenerationError{
				Component: "Entity",
				Stage:     "validation",
				Err:       errors.New("invalid field"),
			},
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapGenerationError(tt.component, tt.stage, tt.err)
			var generationErr *domain.GenerationError
			if errors.As(result, &generationErr) != tt.wantType {
				t.Errorf("WrapGenerationError() type assertion = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestExitCodes(t *testing.T) {
	// Verify exit codes are distinct
	codes := map[int]string{
		ExitCodeSuccess:         "success",
		ExitCodeValidationError: "validation",
		ExitCodeFileSystemError: "filesystem",
		ExitCodeTemplateError:   "template",
		ExitCodeGenerationError: "generation",
		ExitCodeUnknownError:    "unknown",
	}

	seen := make(map[int]bool)
	for code, name := range codes {
		if seen[code] {
			t.Errorf("Duplicate exit code %d for %s", code, name)
		}
		seen[code] = true
	}

	// Verify exit codes are in expected ranges
	if ExitCodeSuccess != 0 {
		t.Errorf("ExitCodeSuccess should be 0, got %d", ExitCodeSuccess)
	}

	if ExitCodeValidationError < 1 || ExitCodeValidationError > 10 {
		t.Errorf("ExitCodeValidationError should be in range 1-10, got %d", ExitCodeValidationError)
	}
}
