---
title: Installation
description: Install scaffold-go from source or go install, then verify the CLI.
---

# Installation

You need a Go toolchain that can build the module declared in
[go.mod](https://github.com/bancolombia/scaffold-go/blob/master/go.mod)
(`go 1.25.0` in this repository). CI currently builds with Go 1.26.x
([.github/workflows/build.yml](https://github.com/bancolombia/scaffold-go/blob/master/.github/workflows/build.yml)).

## Option 1: `go install`

```bash
go install github.com/bancolombia/scaffold-go@latest
```

The binary is named `main` unless you install a versioned module path that
exposes a `main` package under `cmd`. Prefer building from source (below) if
the installed name is not `scaffold-go`.

The Go module path is `github.com/bancolombia/scaffold-go`; the Git remote is
[bancolombia/scaffold-go](https://github.com/bancolombia/scaffold-go).

## Option 2: Build from source (recommended)

```bash
git clone https://github.com/bancolombia/scaffold-go.git
cd scaffold-go
make build
./bin/scaffold-go version
```

`make build` compiles [cmd/main.go](https://github.com/bancolombia/scaffold-go/blob/master/cmd/main.go)
into `bin/scaffold-go` and injects version metadata via `-ldflags`
([Makefile](https://github.com/bancolombia/scaffold-go/blob/master/Makefile)):

| Variable | Meaning |
|----------|---------|
| `Version` | `git describe --tags` or `dev` |
| `BuildDate` | UTC timestamp |
| `GitCommit` | Short SHA |

Put `./bin` on your `PATH`, or copy the binary:

```bash
sudo cp bin/scaffold-go /usr/local/bin/scaffold-go
```

## Option 3: `go run` during development

```bash
go run cmd/main.go version
go run cmd/main.go help generate
```

## Verify

```bash
scaffold-go version
scaffold-go --help
```

Expected `version` output (values vary):

```text
scaffold-go version dev
Build date: 2026-04-08T...
Git commit: abcdef1
```

Source: [internal/commands/version.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/version.go).

## Global flags

Every command inherits these flags from
[NewRootCommand](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/root.go):

| Flag | Default | Effect |
|------|---------|--------|
| `-v`, `--verbose` | off | Zap development logger at debug level |
| `--config` | `$HOME/.scaffold-go.yaml` | Optional Viper config for the **CLI itself**, not generated apps |

Without `-v`, logs are JSON at info level. Use verbose when a generator fails
and you need template or filesystem steps.

## Related pages

| Page | Relationship |
|------|-------------|
| [First project](./first-project.md) | Run `init` after the binary works |
| [CLI reference](../reference/cli.md) | Full command list |
| [Contributing](../contributing.md) | `make test`, `make build` for contributors |
