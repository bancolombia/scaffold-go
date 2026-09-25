package domain

import "fmt"

// ValidationError represents an error that occurs during input validation
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

// Error implements the error interface for ValidationError
func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation failed for %s '%s': %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// FileSystemError represents an error that occurs during file system operations
type FileSystemError struct {
	Operation string
	Path      string
	Err       error
}

// Error implements the error interface for FileSystemError
func (e *FileSystemError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("file system error during %s operation on '%s': %v", e.Operation, e.Path, e.Err)
	}
	return fmt.Sprintf("file system error during %s operation on '%s'", e.Operation, e.Path)
}

// Unwrap returns the underlying error for FileSystemError
func (e *FileSystemError) Unwrap() error {
	return e.Err
}

// TemplateError represents an error that occurs during template processing
type TemplateError struct {
	TemplateName string
	Operation    string
	Err          error
}

// Error implements the error interface for TemplateError
func (e *TemplateError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("template error in '%s' during %s: %v", e.TemplateName, e.Operation, e.Err)
	}
	return fmt.Sprintf("template error in '%s' during %s", e.TemplateName, e.Operation)
}

// Unwrap returns the underlying error for TemplateError
func (e *TemplateError) Unwrap() error {
	return e.Err
}

// GenerationError represents an error that occurs during code generation
type GenerationError struct {
	Component string
	Stage     string
	Err       error
}

// Error implements the error interface for GenerationError
func (e *GenerationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("generation failed for %s at %s stage: %v", e.Component, e.Stage, e.Err)
	}
	return fmt.Sprintf("generation failed for %s at %s stage", e.Component, e.Stage)
}

// Unwrap returns the underlying error for GenerationError
func (e *GenerationError) Unwrap() error {
	return e.Err
}
