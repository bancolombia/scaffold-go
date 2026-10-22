---
title: Database adapters
description: Generate PostgreSQL, MongoDB, or DynamoDB repository implementations.
---

# Database adapters

`generate da database` implements the domain repository port for one entity.
It is a **driven** adapter: the domain does not know which database you chose.

Source:
[NewGenerateDaDatabaseCommand](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/generate_da.go),
[DatabaseAdapterGenerator](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/database_adapter.go).

## Prerequisites

Generate the entity first. Validation fails if these files are missing:

- `internal/domain/{entity}.go`
- `internal/domain/{entity}_repository.go`

Use the **same** `--fields` list you used for the entity so column/document
mappings match.

## Command

```bash
scaffold-go generate da database \
  --type postgres|mongodb|dynamodb \
  --entity EntityName \
  --fields Name:type,...
```

There is **no** positional name. Identity is `--entity`.

| Flag | Short | Required | Default |
|------|-------|----------|---------|
| `--type` | `-t` | no | `postgres` |
| `--entity` | `-e` | yes | — |
| `--fields` | `-f` | yes | — |

```bash
scaffold-go generate da database \
  --type postgres --entity User --fields Name:string,Email:string,Age:int

scaffold-go gen da database -t mongodb -e Product -f Name:string,Price:float64

scaffold-go generate da database \
  --type dynamodb --entity Order --fields UserID:string,Total:float64,Status:string
```

Invalid `--type` values error: must be `postgres`, `mongodb`, or `dynamodb`.

## Files written

For entity `User` and type `postgres`:

| Path | When |
|------|------|
| `internal/drivenadapters/database/postgres_adapter.go` | First postgres adapter only |
| `internal/drivenadapters/database/user_repository.go` | Always (entity-specific) |
| `internal/drivenadapters/database/user_repository_test.go` | Always |
| `internal/drivenadapters/database/module.go` | First database generate only |

Templates:
[templates/da/database](https://github.com/bancolombia/scaffold-go/tree/master/internal/template/templates/da/database).

The repository should satisfy `domain.IUserRepository`. Confirm the Fx module
provides the adapter as that interface (`fx.As` or constructor return type).

## Config YAML

After files are written,
`EnsureConfigBlockForDatabase`
([configwiring.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/configwiring.go))
appends a top-level YAML block if it is not already present.

| Type | Template | Top-level key |
|------|----------|---------------|
| postgres | `config/postgres.yaml.tmpl` | `database` |
| mongodb | `config/mongodb.yaml.tmpl` | `mongodb` |
| dynamodb | `config/dynamodb.yaml.tmpl` | (see template) |

Postgres snippet uses host `localhost`, port `5432`, database name from the
module basename. Prefer `APP_POSTGRES_*` / `APP_*` env vars as documented in
the YAML comments and [configuration](./configuration.md).

If `config/config.yaml` is missing, the CLI warns and skips injection.

## What to do next

1. Align SQL/collection names with your schema (`toTableName` helpers exist
   in the generator package).
2. Provide a real DB client in the adapter constructor (pgx, mongo driver,
   AWS SDK).
3. Register the repository in Fx so use cases receive `I{Entity}Repository`.
4. Never import `drivenadapters/database` from `domain`.

## Related pages

| Page | Relationship |
|------|-------------|
| [Entities](./entities.md) | Port this adapter implements |
| [Configuration](./configuration.md) | YAML blocks and env vars |
| [Init](./init.md) | Empty `drivenadapters/database` folder |
