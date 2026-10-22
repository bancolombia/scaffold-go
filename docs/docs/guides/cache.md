---
title: Cache adapters
description: Generate a cache driven adapter with Get, Set, Delete, and Exists stubs.
---

# Cache adapters

`generate da cache` writes a stub cache you connect to Redis or an in-memory
store. It does not add YAML config automatically.

Source:
[NewGenerateDaCacheCommand](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/generate_da.go),
[CacheGenerator](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/cache.go).

## Command

```bash
scaffold-go generate da cache [name]
```

No extra flags. Name must be PascalCase.

```bash
scaffold-go generate da cache UserCache
scaffold-go gen da cache SessionCache
```

## Files

For `UserCache`:

| Path |
|------|
| `internal/drivenadapters/cache/usercache_cache.go` |
| `internal/drivenadapters/cache/module.go` |

Default TTL in the template is five minutes. Methods: `Get`, `Set`, `Delete`,
`Exists` (see
[cache_adapter.go.tmpl](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/templates/da/cache/cache_adapter.go.tmpl)).

`Get`/`Set` contain commented Redis examples (`go-redis`). Wire a client and
replace the TODOs.

## What to do next

1. Add a domain cache port if use cases should not import `cache`.
2. Inject `*redis.Client` (or similar) in `New{Name}Cache`.
3. Define `ErrCacheMiss` and handle it in use cases.
4. Optionally add a `cache:` block to `config.yaml` yourself (no generator
   hook yet; see the recipe in
   [configwiring.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/configwiring.go)).

## Related pages

| Page | Relationship |
|------|-------------|
| [Database adapters](./database-adapters.md) | Persistence vs cache |
| [Configuration](./configuration.md) | Where Redis URL would live |
| [Use cases](./use-cases.md) | Call cache behind a port |
