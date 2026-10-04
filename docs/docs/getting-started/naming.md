---
title: Naming and validation
description: Project names, component names, and field formats the CLI accepts.
---

# Naming and validation

The CLI rejects invalid names **before** it writes files. Rules live in
[internal/validator/validator.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/validator/validator.go).

## Summary

| What | Rule | Example valid | Example invalid |
|------|------|---------------|-----------------|
| Project name (`init`) | `^[a-z][a-z0-9-]*$` | `blog-service` | `Blog`, `1svc`, `My_App` |
| Component name (entity, usecase, ep, da) | `^[A-Z][a-zA-Z0-9]*$` | `User`, `CreatePost` | `user`, `create-post` |
| Field name | PascalCase, `Name:type` | `Email:string` | `email:string`, `Email` |
| Directory for `init` | Must **not** already exist | new folder | existing path |

## Project names

```bash
scaffold-go init my-service
scaffold-go init my-service --module github.com/acme/my-service
```

If `--module` is omitted, the module path equals the project name
([init.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/init.go)).

## Component names

Used by `entity`, `interface`, `usecase`, `ep http`, `ep consumer`, `da api`,
`da publisher`, and `da cache`:

```text
User          ✓
CreateUser    ✓
User2         ✓
user          ✗ must start with uppercase
create_user   ✗ no underscores
```

HTTP handlers: if the name ends with `Handler`, that suffix is stripped
before generation (`UserHandler` → `User`).

## Fields

Format `Name:type`, comma-separated. Repeatable via Cobra string slices
(`-f a:string -f b:int` also works).

```bash
scaffold-go generate entity Order \
  --fields UserID:string,Total:float64,CreatedAt:time.Time
```

Supported in practice: any Go type string you pass through. `time.Time`
triggers a `time` import on the entity template
([entity.go generator](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/entity.go)).
The validator only checks `Name:type` shape and PascalCase names — it does
not whitelist types.

## Dependencies and use cases

| Flag | Shape | Example |
|------|-------|---------|
| `--dependencies` | `name:Type` | `postRepo:IPostRepository` |
| `--usecases` | `name:Type` or `name:Type:Package` | `createPost:CreatePostUseCase` |
| `--routes` | `METHOD:path:handler` | `POST:/posts:CreatePost` |

Use-case dependency types are imported from `internal/domain`
([usecase.go.tmpl](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/templates/domain/usecase/usecase.go.tmpl)).
Prefer domain interfaces (`IUserRepository`), not adapter structs.

## Project must be initialized

`generate *` fails unless a `go.mod` exists in the current or a parent
directory:

```text
project not initialized: no go.mod file found ...
Run 'scaffold-go init <project-name>' first
```

## Related pages

| Page | Relationship |
|------|-------------|
| [Init](../guides/init.md) | Project name and `--output` |
| [Entities](../guides/entities.md) | Field list in generated structs |
| [Troubleshooting](../reference/troubleshooting.md) | Common validation errors |
