---
title: Contributing
description: Develop, test, and document changes to scaffold-go.
---

# Contributing

Bug reports, docs, and code are welcome. Discuss large features in an issue
first. Full policy:
[CONTRIBUTING.md](https://github.com/bancolombia/scaffold-go/blob/master/CONTRIBUTING.md).

## Setup

Requires the Go version in
[go.mod](https://github.com/bancolombia/scaffold-go/blob/master/go.mod).

```bash
git clone https://github.com/bancolombia/scaffold-go.git
cd scaffold-go
git switch -c your-change
make test-fast
make build
```

## Make targets

From the [Makefile](https://github.com/bancolombia/scaffold-go/blob/master/Makefile):

| Target | Purpose |
|--------|---------|
| `make build` | `bin/scaffold-go` with version ldflags |
| `make test-fast` | Short tests |
| `make test-unit` | Unit tests (`-short`) |
| `make test-integration` | `tests/integration` |
| `make test-properties` | Gopter with fewer iterations |
| `make test` | Full suite |
| `make test-coverage` | `coverage.html` |
| `make fmt` | `go fmt ./...` |

When you change generated output, add or update tests that assert file
contents (generator tests and
[tests/integration/full_workflow_test.go](https://github.com/bancolombia/scaffold-go/blob/master/tests/integration/full_workflow_test.go)).

## Docs site

Markdown is in `/docs/docs`. Preview:

```bash
cd docs
npm install
npm start
```

See [Publishing docs](./publishing.md) for GitHub Pages.

## Pull requests

- Default branch is `master`.
- Summarize the problem, link issues, list checks you ran.
- Update these docs when commands or flags change.
- License: Apache 2.0 ([LICENCE](https://github.com/bancolombia/scaffold-go/blob/master/LICENCE)).

Security reports: [SECURITY.md](https://github.com/bancolombia/scaffold-go/blob/master/SECURITY.md).

## Related pages

| Page | Relationship |
|------|-------------|
| [CLI internals](./architecture/cli-internals.md) | Where to change generators |
| [Publishing docs](./publishing.md) | Docusaurus + GitHub Pages |
