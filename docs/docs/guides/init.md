---
title: Initialize a project
description: scaffold-go init flags, directories, and files created from templates.
---

# Initialize a project

`scaffold-go init` is the only command that does **not** require an existing
Go module. It creates the tree later generators assume.

Source: [internal/commands/init.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/init.go)
(`NewInitCommand`, `initProject`).

## Command

```bash
scaffold-go init [project-name] [flags]
```

| Flag | Short | Default | Purpose |
|------|-------|---------|---------|
| `--module` | `-m` | project name | Go module path written to `go.mod` |
| `--output` | `-o` | `.` | Parent directory; project folder is `{output}/{name}` |

Examples:

```bash
scaffold-go init my-service
scaffold-go init my-service --module github.com/user/my-service
scaffold-go init my-service -m github.com/user/my-service -o ./projects
```

Validation:

- Project name: lowercase, digits, hyphens ([naming](../getting-started/naming.md))
- Target directory must not exist (`ValidateDirectoryNotExists`)

## Directories created

| Path | Role |
|------|------|
| `cmd/` | Application `main` |
| `config/` | Viper config struct + YAML |
| `internal/domain/` | Entities and ports |
| `internal/entrypoints/http/` | Fiber handlers (generated later) |
| `internal/entrypoints/messaging/` | Consumers |
| `internal/drivenadapters/database/` | DB adapters |
| `internal/drivenadapters/api/` | HTTP clients |
| `internal/drivenadapters/messaging/` | Publishers |
| `internal/drivenadapters/cache/` | Cache |
| `internal/infrastructure/messaging/` | Shared broker Fx modules |

## Files from templates

Templates under
[internal/template/templates/project/](https://github.com/bancolombia/scaffold-go/tree/master/internal/template/templates/project):

| Template | Output |
|----------|--------|
| `go.mod.tmpl` | `go.mod` |
| `main.go.tmpl` | `cmd/main.go` |
| `config.go.tmpl` | `config/config.go` |
| `config.yaml.tmpl` | `config/config.yaml` |
| `README.md.tmpl` | `README.md` |
| `Makefile.tmpl` | `Makefile` |

`cmd/main.go` includes **scaffold markers** that later commands patch:

```go
// scaffold-go:imports:begin
"github.com/user/my-service/config"
// scaffold-go:imports:end

app := fx.New(
    fx.Invoke(run),
    fx.Provide(
        config.NewConfig,
        zap.NewProduction,
    ),
    // scaffold-go:modules:begin
    // scaffold-go:modules:end
)
```

Source:
[project/main.go.tmpl](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/templates/project/main.go.tmpl).

Do not delete those comments. Consumers and publishers insert imports and
`fx.Module` lines between them.

## After init

The CLI prints:

```text
Next steps:
  cd <projectDir>
  go mod tidy
  scaffold-go generate entity User --fields Name:string,Email:string
```

`go mod tidy` is required: generated `go.mod` lists a module path but
dependencies (Fx, Zap, Viper, Fiber, drivers) appear as you generate
components and edit code.

Default YAML:

```yaml
server:
  host: localhost
  port: 8080

logging:
  level: info
```

[config.yaml.tmpl](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/templates/project/config.yaml.tmpl).
Database blocks are appended later ([Configuration](./configuration.md)).

## Related pages

| Page | Relationship |
|------|-------------|
| [Generated project](../architecture/generated-project.md) | Layout after several generate commands |
| [Entities](./entities.md) | Next command |
| [Dependency injection](./dependency-injection.md) | Marker-based wiring |
