package commands

import (
	"go.uber.org/zap"
)

// LogCommandStart logs the start of a command execution with context
func LogCommandStart(logger *zap.Logger, commandType, componentName string, fields ...zap.Field) {
	baseFields := []zap.Field{
		zap.String("command", commandType),
		zap.String("component_name", componentName),
	}
	baseFields = append(baseFields, fields...)
	logger.Info("Starting command execution", baseFields...)
}

// LogCommandSuccess logs successful command completion with context
func LogCommandSuccess(logger *zap.Logger, commandType, componentName string, fields ...zap.Field) {
	baseFields := []zap.Field{
		zap.String("command", commandType),
		zap.String("component_name", componentName),
	}
	baseFields = append(baseFields, fields...)
	logger.Info("Command completed successfully", baseFields...)
}

// LogValidationStart logs the start of validation
func LogValidationStart(logger *zap.Logger, validationType string) {
	logger.Debug("Starting validation",
		zap.String("validation_type", validationType),
	)
}

// LogValidationError logs validation errors with context
func LogValidationError(logger *zap.Logger, validationType string, err error, fields ...zap.Field) {
	baseFields := []zap.Field{
		zap.String("validation_type", validationType),
		zap.Error(err),
	}
	baseFields = append(baseFields, fields...)
	logger.Error("Validation failed", baseFields...)
}

// LogGenerationStart logs the start of code generation
func LogGenerationStart(logger *zap.Logger, componentType string) {
	logger.Debug("Starting code generation",
		zap.String("component_type", componentType),
	)
}

// LogGenerationError logs generation errors with context
func LogGenerationError(logger *zap.Logger, componentType string, err error) {
	logger.Error("Code generation failed",
		zap.String("component_type", componentType),
		zap.Error(err),
	)
}

// LogFileOperation logs file system operations
func LogFileOperation(logger *zap.Logger, operation, path string) {
	logger.Debug("File operation",
		zap.String("operation", operation),
		zap.String("path", path),
	)
}

// LogFileOperationError logs file operation errors
func LogFileOperationError(logger *zap.Logger, operation, path string, err error) {
	logger.Error("File operation failed",
		zap.String("operation", operation),
		zap.String("path", path),
		zap.Error(err),
	)
}

// LogTemplateProcessing logs template processing operations
func LogTemplateProcessing(logger *zap.Logger, templateName, outputPath string) {
	logger.Debug("Processing template",
		zap.String("template", templateName),
		zap.String("output", outputPath),
	)
}

// LogTemplateError logs template processing errors
func LogTemplateError(logger *zap.Logger, templateName string, err error) {
	logger.Error("Template processing failed",
		zap.String("template", templateName),
		zap.Error(err),
	)
}
