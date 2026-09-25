package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/gabheadz/scaffold-go/internal/domain"
	"go.uber.org/zap"
)

// Exit codes for different error types
const (
	ExitCodeSuccess         = 0
	ExitCodeValidationError = 1
	ExitCodeFileSystemError = 2
	ExitCodeTemplateError   = 3
	ExitCodeGenerationError = 4
	ExitCodeUnknownError    = 99
)

// HandleError processes errors and returns appropriate exit codes
func HandleError(err error, logger *zap.Logger) int {
	if err == nil {
		return ExitCodeSuccess
	}

	var validationErr *domain.ValidationError
	var fileSystemErr *domain.FileSystemError
	var templateErr *domain.TemplateError
	var generationErr *domain.GenerationError

	switch {
	case errors.As(err, &validationErr):
		logger.Error("Validation error", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return ExitCodeValidationError

	case errors.As(err, &fileSystemErr):
		logger.Error("File system error", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return ExitCodeFileSystemError

	case errors.As(err, &templateErr):
		logger.Error("Template error", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return ExitCodeTemplateError

	case errors.As(err, &generationErr):
		logger.Error("Generation error", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return ExitCodeGenerationError

	default:
		logger.Error("Unknown error", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return ExitCodeUnknownError
	}
}

// WrapValidationError wraps an error as a ValidationError if it isn't already
func WrapValidationError(field, value string, err error) error {
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		return err
	}
	return &domain.ValidationError{
		Field:   field,
		Value:   value,
		Message: err.Error(),
	}
}

// WrapFileSystemError wraps an error as a FileSystemError if it isn't already
func WrapFileSystemError(operation, path string, err error) error {
	var fileSystemErr *domain.FileSystemError
	if errors.As(err, &fileSystemErr) {
		return err
	}
	return &domain.FileSystemError{
		Operation: operation,
		Path:      path,
		Err:       err,
	}
}

// WrapTemplateError wraps an error as a TemplateError if it isn't already
func WrapTemplateError(templateName, operation string, err error) error {
	var templateErr *domain.TemplateError
	if errors.As(err, &templateErr) {
		return err
	}
	return &domain.TemplateError{
		TemplateName: templateName,
		Operation:    operation,
		Err:          err,
	}
}

// WrapGenerationError wraps an error as a GenerationError if it isn't already
func WrapGenerationError(component, stage string, err error) error {
	var generationErr *domain.GenerationError
	if errors.As(err, &generationErr) {
		return err
	}
	return &domain.GenerationError{
		Component: component,
		Stage:     stage,
		Err:       err,
	}
}
