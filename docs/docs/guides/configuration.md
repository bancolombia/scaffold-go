---
title: Configuration
description: Viper config, YAML files, environment variables, and database blocks.
---

# Configuration

Generated services load config with Viper. The CLI's own `--config` flag is
unrelated (it reads `$HOME/.scaffold-go.yaml` for the **generator**).

## Files

| File | Role |
|------|------|
| `config/config.go` | Structs + `NewConfig` |
| `config/config.yaml` | Defaults committed to the repo |

Source:
[config.go.tmpl](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/templates/project/config.go.tmpl).

`NewConfig`:

- File name `config`, type `yaml`
- Search paths `./config` and `.`
- Env prefix `APP`, dots become underscores (`server.port` → `APP_SERVER_PORT`)
- Unmarshal into `Config`

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    MongoDB  MongoDBConfig
    DynamoDB DynamoDBConfig
    Logging  LoggingConfig
}
```

## Default YAML

```yaml
server:
  host: localhost
  port: 8080

logging:
  level: info
```

Override without editing the file:

```bash
export APP_SERVER_PORT=9090
export APP_LOGGING_LEVEL=debug
```

## Database blocks

`generate da database` appends YAML if the top-level key is absent
([EnsureConfigBlock](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/configwiring.go)).

Postgres (`config/postgres.yaml.tmpl`):

```yaml
database:
  host: localhost
  port: 5432
  user: # or APP_POSTGRES_USER / env mapping in comments
  password:
  database: <module-basename>
```

MongoDB (`config/mongodb.yaml.tmpl`):

```yaml
mongodb:
  uri: mongodb://%s:%s@localhost:27017
  database: <module-basename>
  user:
  password:
```

Re-running the generator is idempotent for these keys (`hasTopLevelKey`).

:::warning Secrets
Do not commit real passwords. Leave YAML empty and use environment variables.
:::

## Extending config

The comment on `EnsureConfigBlock` describes the recipe:

1. Add `internal/template/templates/config/<key>.yaml.tmpl`
2. Call `EnsureConfigBlock` after `WriteFiles`
3. Test presence + idempotency

Cache and API generators do not call this yet — add keys to `config.go` and
YAML manually.

## Related pages

| Page | Relationship |
|------|-------------|
| [Init](./init.md) | Creates the initial YAML |
| [Database adapters](./database-adapters.md) | Injects DB blocks |
| [CLI internals](../architecture/cli-internals.md) | How wiring helpers work |
