# Error Types Usage Examples

This document provides examples of how to use the custom error types defined in `errors.go`.

## ValidationError

Use `ValidationError` when input validation fails:

```go
// Example 1: Invalid component name
if !isValidName(name) {
    return &domain.ValidationError{
        Field:   "name",
        Value:   name,
        Message: "must start with uppercase letter and contain only alphanumeric characters",
    }
}

// Example 2: Missing required field
if fields == nil || len(fields) == 0 {
    return &domain.ValidationError{
        Field:   "fields",
        Value:   "",
        Message: "at least one field is required",
    }
}
```

## FileSystemError

Use `FileSystemError` when file system operations fail:

```go
// Example 1: Failed to create directory
if err := os.MkdirAll(path, 0755); err != nil {
    return &domain.FileSystemError{
        Operation: "create directory",
        Path:      path,
        Err:       err,
    }
}

// Example 2: Failed to write file
if err := os.WriteFile(filePath, content, 0644); err != nil {
    return &domain.FileSystemError{
        Operation: "write file",
        Path:      filePath,
        Err:       err,
    }
}
```

## TemplateError

Use `TemplateError` when template processing fails:

```go
// Example 1: Template execution failed
content, err := templates.Execute("entity.go.tmpl", data)
if err != nil {
    return &domain.TemplateError{
        TemplateName: "entity.go.tmpl",
        Operation:    "execution",
        Err:          err,
    }
}

// Example 2: Template not found
tmpl := templates.Lookup(templateName)
if tmpl == nil {
    return &domain.TemplateError{
        TemplateName: templateName,
        Operation:    "lookup",
        Err:          nil,
    }
}
```

## GenerationError

Use `GenerationError` when code generation fails:

```go
// Example 1: Validation stage failed
if err := generator.Validate(); err != nil {
    return &domain.GenerationError{
        Component: "Entity",
        Stage:     "validation",
        Err:       err,
    }
}

// Example 2: Generation stage failed
files, err := generator.Generate()
if err != nil {
    return &domain.GenerationError{
        Component: "Handler",
        Stage:     "generation",
        Err:       err,
    }
}
```

## Error Wrapping and Unwrapping

The error types support Go's error wrapping:

```go
// Wrapping errors
err := &domain.FileSystemError{
    Operation: "write",
    Path:      "/path/to/file",
    Err:       originalErr,
}

// Unwrapping errors
if unwrapped := errors.Unwrap(err); unwrapped != nil {
    // Handle the underlying error
}

// Using errors.Is and errors.As
if errors.Is(err, os.ErrNotExist) {
    // Handle file not found
}
```

## Integration with Existing Code

To integrate these error types into existing code, replace generic `fmt.Errorf` calls with the appropriate error type:

### Before:
```go
func (g *EntityGenerator) Validate() error {
    if err := g.Validator.ValidateName(g.Name); err != nil {
        return fmt.Errorf("invalid entity name: %w", err)
    }
    return nil
}
```

### After:
```go
func (g *EntityGenerator) Validate() error {
    if err := g.Validator.ValidateName(g.Name); err != nil {
        return &domain.ValidationError{
            Field:   "entity name",
            Value:   g.Name,
            Message: err.Error(),
        }
    }
    return nil
}
```
