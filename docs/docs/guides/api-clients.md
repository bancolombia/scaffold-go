---
title: API clients
description: Generate outbound HTTP clients as driven adapters.
---

# API clients

`generate da api` scaffolds a client for **another** HTTP API. Domain code
should depend on an interface you define; the generated client is the
implementation.

Source:
[NewGenerateDaApiCommand](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/generate_da.go),
[ApiClientGenerator](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/api_client.go).

## Command

```bash
scaffold-go generate da api [name] [--url https://api.example.com]
```

| Flag | Short | Required |
|------|-------|----------|
| `--url` | `-u` | no |

```bash
scaffold-go generate da api PaymentClient --url https://api.payment.com
scaffold-go gen da api WeatherClient -u https://api.weather.com
```

`name` must be PascalCase ([naming](../getting-started/naming.md)).

## Files

For `PaymentClient`:

| Path |
|------|
| `internal/drivenadapters/api/paymentclient_client.go` |
| `internal/drivenadapters/api/module.go` |

`module.go` is overwritten each run (unlike database `module.go`, which
skips if present). If you already customized the module, re-apply your
providers after regenerating.

## What to do next

1. Add methods that match a domain port (`Payments`, `FxRates`, …).
2. Inject `*http.Client` / timeouts / tracing.
3. Map HTTP errors to domain errors; do not leak status codes into use cases.
4. Put base URL in `config` (the `--url` value is template data, not
   automatically added to `config.yaml` today).
5. `go mod tidy` if you add HTTP helper libraries.

## Related pages

| Page | Relationship |
|------|-------------|
| [Use cases](./use-cases.md) | Depend on a domain interface, not the client struct |
| [HTTP entry-points](./http-entry-points.md) | Inbound HTTP vs this outbound client |
| [Configuration](./configuration.md) | Where to store base URLs |
