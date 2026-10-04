---
title: Templates
description: Inventory of embedded Go templates used by scaffold-go generators.
---

# Templates

Templates are embedded from
[internal/template/templates](https://github.com/bancolombia/scaffold-go/tree/master/internal/template/templates)
and executed by
[TemplateEngine](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/engine.go).
Data structs live in
[internal/template/models.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/models.go).

## Project (`init`)

| Template | Output |
|----------|--------|
| `project/go.mod.tmpl` | `go.mod` |
| `project/main.go.tmpl` | `cmd/main.go` |
| `project/config.go.tmpl` | `config/config.go` |
| `project/config.yaml.tmpl` | `config/config.yaml` |
| `project/README.md.tmpl` | `README.md` |
| `project/Makefile.tmpl` | `Makefile` |

## Domain

| Template | Output |
|----------|--------|
| `domain/entity/entity.go.tmpl` | `internal/domain/{name}.go` |
| `domain/entity/repository.go.tmpl` | `internal/domain/{name}_repository.go` |
| `domain/entity/entity_test.go.tmpl` | `internal/domain/{name}_test.go` |
| `domain/interface/interface.go.tmpl` | `internal/domain/{lowercase-name}.go` (`I{Name}`) |
| `domain/usecase/usecase.go.tmpl` | `internal/usecase/{name}_usecase.go` |
| `domain/usecase/usecase_test.go.tmpl` | tests |
| `domain/entity/publisher.go.tmpl` | Domain-side publisher port (when used) |

## HTTP entry-points

| Template | Output |
|----------|--------|
| `ep/http/handler/http_handler.go.tmpl` | `{name}_handler.go` |
| `ep/http/handler/health_handler.go.tmpl` | health |
| `ep/http/handler/module.go.tmpl` | Fx module |
| `ep/http/routes/*.tmpl` | routers |
| `ep/http/middleware/cors.go.tmpl` | CORS |
| `ep/http/server.go.tmpl` | Fiber server |
| `ep/http/handler_test.go.tmpl` | tests |

## Consumers and publishers

Under `ep/consumer/{rabbitmq,ibmmq}/` and
`da/publisher/{rabbitmq,ibmmq}/`, plus
`infrastructure/messaging/{rabbitmq,ibmmq}/module.go.tmpl`.

## Driven adapters

| Area | Templates |
|------|-----------|
| Database | `{postgres,mongodb,dynamodb}_adapter.go.tmpl`, `*_repository.go.tmpl`, tests, `database_module.go.tmpl` |
| API | `da/api/api_client.go.tmpl`, `api_module.go.tmpl` |
| Cache | `da/cache/cache_adapter.go.tmpl`, `cache_module.go.tmpl` |
| Config snippets | `config/{postgres,mongodb,dynamodb}.yaml.tmpl` |

## Helpers

Templates use functions such as `toLower`, `toCamelCase` registered on the
engine. Keep new templates valid Go **after** execution so `gofmt` succeeds.

## Related pages

| Page | Relationship |
|------|-------------|
| [CLI internals](../architecture/cli-internals.md) | How templates are executed |
| [Contributing](../contributing.md) | Changing generated output |
