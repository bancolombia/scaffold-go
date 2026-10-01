---
title: Troubleshooting
description: Common scaffold-go errors and how to fix them.
---

# Troubleshooting

Run the failing command with `-v` to see validation, template, and write
steps ([root.go logger](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/root.go)).

## Project not initialized

```text
project not initialized: no go.mod file found ...
```

You are not inside an `init`'d module (or any Go module). `cd` into the
project or run `scaffold-go init` first
([ValidateProjectInitialized](https://github.com/bancolombia/scaffold-go/blob/master/internal/validator/validator.go)).

## Invalid project or component name

Project names: lowercase + hyphens (`my-service`). Component names:
PascalCase (`CreateUser`). Fields: `Name:type`.

See [Naming](../getting-started/naming.md).

## Directory already exists

`init` refuses to overwrite. Pick another name or `--output`, or remove the
folder.

## Entity not found (database)

```text
entity "User" not found at internal/domain/user.go
```

Run `generate entity` before `generate da database`. The repository file
must exist too.

## Kafka

`--broker kafka` is rejected with a friendly error. Use `rabbitmq` or
`ibmmq`.

## IBM MQ / CGO

Generated IBM MQ code depends on
[commons-mq-go](https://github.com/bancolombia/commons-mq-go) and the IBM MQ
C client. Install the client and enable CGO before `go build`.

## main.go not patched

Delete or edit of `// scaffold-go:imports:begin` / `modules` markers prevents
auto-wiring. Restore the comments from
[main.go.tmpl](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/templates/project/main.go.tmpl)
or add imports manually.

Module path resolution failure prints:

```text
could not resolve module path; skipping cmd/main.go auto-wiring
```

## Fx missing type

Startup error that no constructor provides `IUserRepository` (or similar):
register the database adapter as that interface. Second HTTP handlers may
need a manual `fx.Provide(NewFooHandler)` ([HTTP guide](../guides/http-entry-points.md)).

## gofmt / template errors

If generation fails while formatting, the template emitted invalid Go.
Fix the template or the data (empty module path in tests is expected).

## `--force` / `--output-path` ignored

Those flags are declared on `generate` but generators use `projectDir := "."`.
Run commands from the module root.

## Related pages

| Page | Relationship |
|------|-------------|
| [CLI reference](./cli.md) | Flag list |
| [Naming](../getting-started/naming.md) | Validation regexes |
| [Dependency injection](../guides/dependency-injection.md) | Wiring failures |
