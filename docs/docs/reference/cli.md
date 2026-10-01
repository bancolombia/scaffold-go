---
title: CLI reference
description: Commands, aliases, and flags for scaffold-go.
---

# CLI reference

Binary entry: [cmd/main.go](https://github.com/bancolombia/scaffold-go/blob/master/cmd/main.go)
→ [NewRootCommand](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/root.go).

## Global

```bash
scaffold-go [command] [-v|--verbose] [--config path]
```

| Flag | Description |
|------|-------------|
| `-v`, `--verbose` | Debug Zap logs |
| `--config` | CLI Viper file (default `$HOME/.scaffold-go.yaml`) |

## `version`

```bash
scaffold-go version
```

Prints `Version`, `BuildDate`, `GitCommit`
([version.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/version.go)).

## `init`

```bash
scaffold-go init [project-name] [-m|--module path] [-o|--output dir]
```

See [Initialize a project](../guides/init.md).

## `generate` (`gen`, `g`)

```bash
scaffold-go generate <subcommand>
scaffold-go gen entity ...
```

Persistent flags (declared, not consumed by generators today):

| Flag | Description |
|------|-------------|
| `-o`, `--output-path` | Intended output directory |
| `--force` | Intended overwrite |

### `generate entity`

```bash
scaffold-go generate entity [name] -f|--fields Name:type,...
```

### `generate usecase`

```bash
scaffold-go generate usecase [name] [-d|--dependencies name:Type,...]
```

### `generate ep`

Subcommands: `http`, `consumer`.

```bash
scaffold-go generate ep http [name] \
  [-u|--usecases name:Type] \
  [-r|--routes METHOD:path:Handler] \
  [--enable-cors] [--enable-rate-limiter] [--enable-error-handler]
```

```bash
scaffold-go generate ep consumer [name] \
  -b|--broker rabbitmq|ibmmq|kafka \
  -p|--pattern <pattern> \
  [-u|--usecases name:Type[:Package]] \
  [-q|--queue name]
```

### `generate da`

Subcommands: `database`, `api`, `publisher`, `cache`.

```bash
scaffold-go generate da database \
  [-t|--type postgres|mongodb|dynamodb] \
  -e|--entity Name \
  -f|--fields Name:type,...
```

```bash
scaffold-go generate da api [name] [-u|--url URL]
scaffold-go generate da publisher [name] -b|--broker ... -p|--pattern ... [-t|--target ...]
scaffold-go generate da cache [name]
```

## Help

```bash
scaffold-go help
scaffold-go help generate
scaffold-go generate entity --help
```

## Related pages

| Page | Relationship |
|------|-------------|
| [Naming](../getting-started/naming.md) | Validation rules |
| [Troubleshooting](./troubleshooting.md) | Error messages |
| [First project](../getting-started/first-project.md) | Combined example |
