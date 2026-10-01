---
title: Feature workflow
description: Recommended order for adding a feature with scaffold-go generate.
---

# Feature workflow

`generate` documents the order in its long help
([internal/commands/generate.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/generate.go)):

1. **entity** — model + repository port
2. **usecase** — application logic that depends on ports
3. **ep** — HTTP or message consumer that calls the use case
4. **da** — database, API, publisher, or cache that implements ports

Skipping entity before `da database` fails validation: the adapter generator
checks that `internal/domain/{entity}.go` and `{entity}_repository.go` exist
([database_adapter.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/database_adapter.go)).

## Dependency rule

```mermaid
graph TB
  subgraph outside ["Outside"]
    HTTP["HTTP / Fiber"]
    MQ["Consumers / publishers"]
    DB["Postgres / Mongo / Dynamo"]
  end
  subgraph ep ["EntryPoints"]
    H["Handlers"]
    C["Consumers"]
  end
  subgraph app ["Use cases"]
    UC["CreateX / GetX"]
  end
  subgraph domain ["Domain"]
    E["Entities"]
    I["IRepository / ports"]
  end
  subgraph da ["Driven adapters"]
    R["Repository impls"]
    P["Publishers / API / cache"]
  end
  HTTP --> H
  MQ --> C
  H --> UC
  C --> UC
  UC --> I
  UC --> E
  R --> I
  R --> DB
  P --> MQ
  style HTTP fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style MQ fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style DB fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style H fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style C fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style UC fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style E fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style I fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style R fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
  style P fill:#2d333b,stroke:#6d5dfc,color:#e6edf3
```

<!-- Sources: internal/commands/root.go, internal/commands/generate.go -->

Dependencies point **inward**. Domain templates do not import Fiber, SQL
drivers, or broker SDKs
([entity property tests](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/entity_properties_test.go)).

## Checklist for a new capability

1. `cd` into the initialized module.
2. Generate the entity with the fields you need for persistence **and** HTTP.
3. Generate one use case per action (`CreateX`, `GetX`, `ListX`, …).
4. Choose an entry-point:
   - Synchronous API → [HTTP](./http-entry-points.md)
   - Queue/topic → [consumer](./consumers.md)
5. Implement outbound I/O:
   - Persistence → [database](./database-adapters.md)
   - Other HTTP API → [API client](./api-clients.md)
   - Emit messages → [publisher](./publishers.md)
   - Cache → [cache](./cache.md)
6. `go mod tidy`, fill TODOs, wire any missing Fx providers.
7. Run with `-v` on the CLI if generation skipped files or failed validation.

## Persistent generate flags

Declared on `generate` but **not read** by current generators (they always
use `projectDir := "."`):

| Flag | Declared meaning |
|------|------------------|
| `-o`, `--output-path` | Output directory relative to project root |
| `--force` | Overwrite existing files |

Work from the module root. Re-running some generators is idempotent for
shared files (health handler, CORS, database adapter connection file) and
overwrites component-specific files without `--force`.

## Related pages

| Page | Relationship |
|------|-------------|
| [First project](../getting-started/first-project.md) | Concrete commands |
| [Architecture overview](../architecture/overview.md) | Layer responsibilities |
| [Dependency injection](./dependency-injection.md) | How `cmd/main.go` is patched |
