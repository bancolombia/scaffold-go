---
title: CLI internals
description: How Cobra commands validate input, run generators, format Go, and write files.
---

# CLI internals

This page is for contributors changing scaffold-go itself.

## Pipeline

```mermaid
flowchart LR
  A["Cobra RunE"] --> B["ValidateProjectInitialized"]
  B --> C["NewTemplateEngine"]
  C --> D["generator.Validate"]
  D --> E["generator.Generate"]
  E --> F["FormatGoCode"]
  F --> G["WriteFiles"]
  G --> H["Optional main/config wiring"]
  style A fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style B fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style C fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style D fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style E fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style F fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style G fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style H fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
```

<!-- Sources: internal/commands/generate_entity.go, internal/filesystem/filesystem.go, internal/template/engine.go -->

1. **Command** parses flags ([internal/commands](https://github.com/bancolombia/scaffold-go/tree/master/internal/commands)).
2. **Validator** checks names and `go.mod`.
3. **TemplateEngine** loads embedded `*.tmpl` files.
4. **Generator** returns `[]GeneratedFile`.
5. **Filesystem** runs `gofmt` on `.go` content and writes.
6. **Manager / wiring** patches `cmd/main.go` or `config.yaml`.

Errors are wrapped (`WrapValidationError`, `WrapTemplateError`, …) in
[internal/commands/errors.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/errors.go)
and mapped to exit codes in `cmd/main.go`.

## Adding a new generate kind

1. Add a Cobra command under `generate` (sibling of `entity` / `interface`)
   or under `generate ep` / `generate da`.
2. Implement a generator with `Validate` + `Generate`.
3. Add templates under `internal/template/templates/`.
4. Register templates in the engine if it uses an explicit map
   ([internal/template/engine.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/engine.go)).
5. Unit + property tests next to the generator; integration coverage in
   [tests/integration](https://github.com/bancolombia/scaffold-go/tree/master/tests/integration).
6. Document the command in this site and the root README.

## Tests

| Target | What |
|--------|------|
| `make test-unit` | `go test -short` |
| `make test-properties` | Gopter property tests |
| `make test-integration` | Full CLI workflow |
| `make test` | Everything |

Property tests live beside generators (`*_properties_test.go`) and under
[tests/properties](https://github.com/bancolombia/scaffold-go/tree/master/tests/properties).

## Related pages

| Page | Relationship |
|------|-------------|
| [Templates](../reference/templates.md) | Template inventory |
| [Contributing](../contributing.md) | PR and `make` targets |
| [Architecture overview](./overview.md) | Package map |
