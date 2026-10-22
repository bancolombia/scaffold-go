---
title: Message publishers
description: Generate RabbitMQ or IBM MQ publishers and wire broker modules.
---

# Message publishers

Publishers are **driven** adapters: use cases call a domain port; the
publisher talks to the broker. Pair them with [consumers](./consumers.md) in
other services.

Source:
[NewGenerateDaPublisherCommand](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/generate_da.go),
[internal/generator/publisher.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/publisher.go).

## Command

```bash
scaffold-go generate da publisher [name] \
  --broker rabbitmq|ibmmq|kafka \
  --pattern <pattern> \
  [--target <event|command|resource|queue>]
```

| Flag | Short | Required | Notes |
|------|-------|----------|-------|
| `--broker` | `-b` | yes | `rabbitmq`, `ibmmq`, `kafka` |
| `--pattern` | `-p` | yes | See table |
| `--target` | `-t` | no | Defaults to lowercased name |

Kafka is not implemented (same as consumers).

## Brokers and patterns

| Broker | Patterns | Typical `--target` |
|--------|----------|-------------------|
| `rabbitmq` | `event`, `command`, `query`, `notification` | Event name, command name, resource, app name |
| `ibmmq` | `sender`, `requestreply` | Queue / resource name |

```bash
scaffold-go generate da publisher OrderEventPublisher \
  --broker rabbitmq --pattern event \
  --target order.created

scaffold-go gen da publisher SendInvoicePublisher \
  --broker rabbitmq --pattern command \
  --target send-invoice

scaffold-go generate da publisher InvoiceSender \
  --broker ibmmq --pattern sender \
  --target DEV.QUEUE.INVOICES
```

## What is generated

- Publisher under `internal/drivenadapters/messaging/{broker}/`
- Per-component Fx module
- Shared `internal/infrastructure/messaging/{broker}` module on first use
- `cmd/main.go` patches: foundation import + `publishers.Module`
  (`wirePublisherIntoMain`)

Then run `go mod tidy`.

## What to do next

1. Introduce a domain interface (`OrderEvents`, `IEventBus`) if use cases
   should not import the publisher package.
2. Implement mapping from domain events to broker payloads.
3. IBM MQ: CGO and the IBM MQ C client, same as consumers.
4. Keep `scaffold-go:imports` / `scaffold-go:modules` markers in `main.go`.

## Related pages

| Page | Relationship |
|------|-------------|
| [Consumers](./consumers.md) | Inbound side of the same brokers |
| [Dependency injection](./dependency-injection.md) | Auto-wiring |
| [Templates](../reference/templates.md) | Publisher template paths |
