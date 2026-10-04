---
title: Generate domain interfaces
description: Create a domain port that operates on an existing entity.
---

# Generate domain interfaces

A domain interface is a gateway or port: a contract that use cases depend on
and driven adapters implement. Unlike `generate entity`, this command does
**not** create a struct or a CRUD repository. It declares one operation that
receives an existing entity.

Source:
[internal/commands/generate_interface.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/generate_interface.go),
[internal/generator/interface.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/interface.go).

## Command

```bash
scaffold-go generate interface [name] --entity Entity --operation Operation
```

| Flag | Short | Required | Format |
|------|-------|----------|--------|
| `--entity` | `-e` | yes | PascalCase entity name |
| `--operation` | — | yes | PascalCase method name |

```bash
scaffold-go generate interface ProductReceiver --entity Product --operation ProcessProduct
scaffold-go gen interface OrderGateway -e Order --operation SubmitOrder
```

The positional name must be PascalCase and must **not** include the `I`
prefix. The generator adds it (`ProductReceiver` → `IProductReceiver`).

The entity file is not checked. Generation still succeeds if
`internal/domain/product.go` is missing; the interface simply refers to
`*Product` in the same package.

## Files written

For name `ProductReceiver`:

| File | Content |
|------|---------|
| `internal/domain/productreceiver.go` | `IProductReceiver` |

File names use the same lowercase convention as other generators
(`strings.ToLower`), not snake_case.

## Generated shape

From
[interface.go.tmpl](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/templates/domain/interface/interface.go.tmpl):

```go
package domain

import "context"

// IProductReceiver defines domain operations for Product.
type IProductReceiver interface {
    ProcessProduct(ctx context.Context, product *Product) error
}
```

The parameter name is the entity name with a lowercase first letter.
The signature matches other domain ports: `context.Context`, a pointer to
the entity, and `error`.

## What this command does not do

- It does not generate a test, mock, or driven adapter.
- It does not update `module.go` or `cmd/main.go`.
- It does not accept more than one operation.

Inject the interface from a use case with `--dependencies`, for example
`productReceiver:IProductReceiver`.

## Related pages

| Page | Relationship |
|------|-------------|
| [Entities](./entities.md) | Struct the operation refers to |
| [Use cases](./use-cases.md) | Depend on the generated port |
| [Feature workflow](./workflow.md) | Where this step sits |
